package fault

import (
	"errors"
	"math/rand"
	"net/http"
	"sync"
)

const (
	// defaultRandSeed is used when a random seed is not set explicitly.
	defaultRandSeed = 1
)

var (
	// ErrNilInjector when a nil Injector is passed.
	ErrNilInjector = errors.New("injector cannot be nil")
	// ErrInvalidPercent when a percent is outside of [0.0,1.0).
	ErrInvalidPercent = errors.New("percent must be 0.0 <= percent <= 1.0")
)

// Fault combines an Injector with options on when to use that Injector.
type Fault struct {
	// enabled determines if the fault should evaluate.
	enabled bool

	// injector is the Injector that will be injected.
	injector Injector

	// participation is the percent of requests that run the injector. 0.0 <= p <= 1.0.
	participation float32

	// pathBlocklist is a map of paths that the Injector will never run against.
	pathBlocklist map[string]bool

	// pathAllowlist, if set, is a map of the only paths that the Injector will run against.
	pathAllowlist map[string]bool

	// headerBlocklist is a map of headers that the Injector will never run against.
	headerBlocklist map[string]string

	// headerAllowlist, if set, is a map of the only headers the Injector will run against.
	headerAllowlist map[string]string

	// randSeed is a number to seed rand with.
	randSeed int64

	// rand is our random number source.
	rand *rand.Rand

	// randF is a function that returns a float32 [0.0,1.0).
	randF func() float32

	// randMtx protects Fault.rand, which is not thread safe.
	randMtx sync.Mutex
}

// Option configures a Fault.
type Option interface {
	applyFault(f *Fault) error
}

type enabledOption bool

func (o enabledOption) applyFault(f *Fault) error {
	f.enabled = bool(o)
	return nil
}

// WithEnabled sets if the Fault should evaluate.
func WithEnabled(e bool) Option {
	return enabledOption(e)
}

type participationOption float32

func (o participationOption) applyFault(f *Fault) error {
	if o < 0.0 || o > 1.0 {
		return ErrInvalidPercent
	}
	f.participation = float32(o)
	return nil
}

// WithParticipation sets the percent of requests that run the Injector. 0.0 <= p <= 1.0.
func WithParticipation(p float32) Option {
	return participationOption(p)
}

type pathBlocklistOption []string

func (o pathBlocklistOption) applyFault(f *Fault) error {
	blocklist := make(map[string]bool, len(o))
	for _, path := range o {
		blocklist[path] = true
	}
	f.pathBlocklist = blocklist
	return nil
}

// WithPathBlocklist is a list of paths that the Injector will not run against.
func WithPathBlocklist(blocklist []string) Option {
	return pathBlocklistOption(blocklist)
}

type pathAllowlistOption []string

func (o pathAllowlistOption) applyFault(f *Fault) error {
	allowlist := make(map[string]bool, len(o))
	for _, path := range o {
		allowlist[path] = true
	}
	f.pathAllowlist = allowlist
	return nil
}

// WithPathAllowlist is, if set, a list of the only paths that the Injector will run against.
func WithPathAllowlist(allowlist []string) Option {
	return pathAllowlistOption(allowlist)
}

type headerBlocklistOption map[string]string

func (o headerBlocklistOption) applyFault(f *Fault) error {
	blocklist := make(map[string]string, len(o))
	for key, val := range o {
		blocklist[key] = val
	}
	f.headerBlocklist = blocklist
	return nil
}

// WithHeaderBlocklist is a map of header keys to values that the Injector will not run against.
func WithHeaderBlocklist(blocklist map[string]string) Option {
	return headerBlocklistOption(blocklist)
}

type headerAllowlistOption map[string]string

func (o headerAllowlistOption) applyFault(f *Fault) error {
	allowlist := make(map[string]string, len(o))
	for key, val := range o {
		allowlist[key] = val
	}
	f.headerAllowlist = allowlist
	return nil
}

// WithHeaderAllowlist is, if set, a map of header keys to values of the only headers that the
// Injector will run against.
func WithHeaderAllowlist(allowlist map[string]string) Option {
	return headerAllowlistOption(allowlist)
}

// RandSeedOption configures things that can set a random seed.
type RandSeedOption interface {
	Option
	RandomInjectorOption
}

type randSeedOption int64

func (o randSeedOption) applyFault(f *Fault) error {
	f.randSeed = int64(o)
	return nil
}

// WithRandSeed sets the rand.Rand seed for this struct.
func WithRandSeed(s int64) RandSeedOption {
	return randSeedOption(s)
}

type randFloat32FuncOption func() float32

func (o randFloat32FuncOption) applyFault(f *Fault) error {
	f.randF = o
	return nil
}

// WithRandFloat32Func sets the function that will be used to randomly get our float value. Default
// rand.Float32. Always returns a float32 between [0.0,1.0) to avoid errors.
func WithRandFloat32Func(f func() float32) Option {
	return randFloat32FuncOption(f)
}

// NewFault sets/validates the Injector and Options and returns a usable Fault.
func NewFault(i Injector, opts ...Option) (*Fault, error) {
	if i == nil {
		return nil, ErrNilInjector
	}

	// set defaults
	f := &Fault{
		injector: i,
		randSeed: defaultRandSeed,
		randF:    nil,
	}

	// apply options
	for _, opt := range opts {
		err := opt.applyFault(f)
		if err != nil {
			return nil, err
		}
	}

	// set seeded rand source and function
	f.rand = rand.New(rand.NewSource(f.randSeed))
	if f.randF == nil {
		f.randF = f.rand.Float32
	}

	return f, nil
}

// Handler determines if the Injector should execute and runs it if so.
func (f *Fault) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// By default faults do not evaluate. Here we go through conditions where faults
		// will evaluate, if everything is configured correctly.
		var shouldEvaluate bool

		shouldEvaluate = f.enabled

		shouldEvaluate = shouldEvaluate && f.checkAllowBlockLists(shouldEvaluate, r)

		// false if not selected for participation
		shouldEvaluate = shouldEvaluate && f.participate()

		// run the injector or pass
		if shouldEvaluate {
			f.injector.Handler(next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// SetEnabled updates the enabled state of the Fault.
func (f *Fault) SetEnabled(o enabledOption) error {
	return o.applyFault(f)
}

// SetParticipation updates the participation percentage of the Fault.
func (f *Fault) SetParticipation(o participationOption) error {
	return o.applyFault(f)
}

// checkAllowBlockLists checks the request against the provided allowlists and blocklists, returning
// true if the request may proceed and false otherwise.
func (f *Fault) checkAllowBlockLists(shouldEvaluate bool, r *http.Request) bool {
	// false if path is in pathBlocklist
	shouldEvaluate = shouldEvaluate && !f.pathBlocklist[r.URL.Path]

	// false if pathAllowlist exists and path is not in it
	if len(f.pathAllowlist) > 0 {
		shouldEvaluate = shouldEvaluate && f.pathAllowlist[r.URL.Path]
	}

	// false if any headers match headerBlocklist
	for key, val := range f.headerBlocklist {
		shouldEvaluate = shouldEvaluate && r.Header.Get(key) != val
	}

	// false if headerAllowlist exists and headers are not in it
	if len(f.headerAllowlist) > 0 {
		for key, val := range f.headerAllowlist {
			shouldEvaluate = shouldEvaluate && (r.Header.Get(key) == val)
		}
	}

	return shouldEvaluate
}

// participate randomly decides (returns true) if the Injector should run based on f.participation.
// Numbers outside of [0.0,1.0] will always return false.
func (f *Fault) participate() bool {
	f.randMtx.Lock()
	rn := f.randF()
	f.randMtx.Unlock()

	if rn < f.participation && f.participation <= 1.0 {
		return true
	}

	return false
}
