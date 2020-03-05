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
	// ErrNilInjector returns when a nil Injector type is passed.
	ErrNilInjector = errors.New("injector cannot be nil")
	// ErrInvalidPercent returns when a provided percent is outside of the allowed bounds.
	ErrInvalidPercent = errors.New("percent must be 0.0 <= percent <= 1.0")
)

// Fault combines an Injector with options on when to use that injector.
type Fault struct {
	// enabled determines if the fault should evaluate.
	enabled bool

	// injector is the Injector that will be injected.
	injector Injector

	// participation is the percent of requests that run the injector. 0.0 <= p <= 1.0.
	participation float32

	// pathBlacklist is a map of paths that the injector will not run against.
	pathBlacklist map[string]bool

	// pathWhitelist, if set, is a map of the only paths that the injector will run against.
	pathWhitelist map[string]bool

	// randSeed is a number to seed rand with.
	randSeed int64

	// rand is our random number source.
	rand *rand.Rand

	// randF is a function that returns a float32 [0.0,1.0)
	randF func() float32

	// *rand.Rand is not thread safe. This mutex protects our random source
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

// WithEnabled determines if the fault should evaluate.
func WithEnabled(e bool) Option {
	return enabledOption(e)
}

type participationOption float32

func (o participationOption) applyFault(f *Fault) error {
	if o < 0 || o > 1.0 {
		return ErrInvalidPercent
	}
	f.participation = float32(o)
	return nil
}

// WithParticipation sets the percent of requests that run the injector. 0.0 <= p <= 1.0.
func WithParticipation(p float32) Option {
	return participationOption(p)
}

type pathBlacklistOption []string

func (o pathBlacklistOption) applyFault(f *Fault) error {
	blacklist := make(map[string]bool, len(o))
	for _, path := range o {
		blacklist[path] = true
	}
	f.pathBlacklist = blacklist
	return nil
}

// WithPathBlacklist is a list of paths that the injector will not run against.
func WithPathBlacklist(blacklist []string) Option {
	return pathBlacklistOption(blacklist)
}

type pathWhitelistOption []string

func (o pathWhitelistOption) applyFault(f *Fault) error {
	whitelist := make(map[string]bool, len(o))
	for _, path := range o {
		whitelist[path] = true
	}
	f.pathWhitelist = whitelist
	return nil
}

// WithPathWhitelist is, if set, a map of the only paths that the injector will run against.
func WithPathWhitelist(whitelist []string) Option {
	return pathWhitelistOption(whitelist)
}

// RandSeedOption configures strtucts that can set a random seed
type RandSeedOption interface {
	Option
	RandomInjectorOption
}

type randSeedOption int64

func (o randSeedOption) applyFault(f *Fault) error {
	f.randSeed = int64(o)
	return nil
}

// WithRandSeed sets the seed for fault.rand
func WithRandSeed(s int64) RandSeedOption {
	return randSeedOption(s)
}

type randFloat32FuncOption func() float32

func (o randFloat32FuncOption) applyFault(f *Fault) error {
	f.randF = o
	return nil
}

// WithRandFloat32Func sets the function that will be used to randomly get our float value. Default
// rand.Float32. Make sure your function always returns a float32 between [0.0,1.0) to avoid errors.
func WithRandFloat32Func(f func() float32) Option {
	return randFloat32FuncOption(f)
}

// NewFault validates and sets the provided options and returns a Fault.
func NewFault(i Injector, opts ...Option) (*Fault, error) {
	if i == nil {
		return nil, ErrNilInjector
	}

	// set the defaults.
	fault := &Fault{
		injector: i,
		randSeed: defaultRandSeed,
		randF:    nil,
	}

	// apply the list of options to fault.
	for _, opt := range opts {
		err := opt.applyFault(fault)
		if err != nil {
			return nil, err
		}
	}

	// set our random source and function with the provided seed.
	fault.rand = rand.New(rand.NewSource(fault.randSeed))
	if fault.randF == nil {
		fault.randF = fault.rand.Float32
	}

	return fault, nil
}

// Handler determines if the Injector should execute and runs it if so.
func (f *Fault) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// By default faults do not evaluate. Here we go through conditions where faults
		// will evaluate, if everything is configured correctly.
		var shouldEvaluate bool

		// enabled is the first check, to prioritize speed when faults are disabled.
		if f.enabled {
			// if path is in blacklist, do not evaluate.
			if _, ok := f.pathBlacklist[r.URL.Path]; !ok {
				// if whitelist exists and path is not in it, do not evaluate.
				if len(f.pathWhitelist) > 0 {
					// if path is in the whitelist, evaluate.
					if _, ok := f.pathWhitelist[r.URL.Path]; ok {
						shouldEvaluate = true
					}
				} else {
					// if whitelist does not exist, evaluate.
					shouldEvaluate = true
				}
			}
		}

		// if all conditions pass, check if we're randomly selected to participate
		if shouldEvaluate {
			shouldEvaluate = f.participate()
		}

		// run the injector if shouldEvaluate
		if shouldEvaluate {
			f.injector.Handler(next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// participate randomly decides (returns true) if the injector should run based on f.participation.
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
