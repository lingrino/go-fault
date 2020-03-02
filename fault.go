package fault

import (
	"errors"
	"math/rand"
	"net/http"
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

	// injector Injector that will be injected.
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
}

// Option configures a Fault.
type Option interface {
	applyFault(f *Fault)
}

type enabledOption bool

func (o enabledOption) applyFault(f *Fault) {
	f.enabled = bool(o)
}

// WithEnabled determines if the fault should evaluate.
func WithEnabled(e bool) Option {
	return enabledOption(e)
}

type participationOption float32

func (o participationOption) applyFault(f *Fault) {
	f.participation = float32(o)
}

// WithParticipation sets the percent of requests that run the injector. 0.0 <= p <= 1.0.
func WithParticipation(p float32) Option {
	return participationOption(p)
}

type pathBlacklistOption []string

func (o pathBlacklistOption) applyFault(f *Fault) {
	blacklist := make(map[string]bool, len(o))
	for _, path := range o {
		blacklist[path] = true
	}
	f.pathBlacklist = blacklist
}

// WithPathBlacklist is a list of paths that the injector will not run against.
func WithPathBlacklist(blacklist []string) Option {
	return pathBlacklistOption(blacklist)
}

type pathWhitelistOption []string

func (o pathWhitelistOption) applyFault(f *Fault) {
	whitelist := make(map[string]bool, len(o))
	for _, path := range o {
		whitelist[path] = true
	}
	f.pathWhitelist = whitelist
}

// WithPathWhitelist is, if set, a map of the only paths that the injector will run against.
func WithPathWhitelist(whitelist []string) Option {
	return pathWhitelistOption(whitelist)
}

type RandSeedOption interface {
	Option
	RandomInjectorOption
}

type randSeedOption int64

func (o randSeedOption) applyFault(f *Fault) {
	f.randSeed = int64(o)
}

// WithRandSeed sets the seed for fault.rand
func WithRandSeed(s int64) RandSeedOption {
	return randSeedOption(s)
}

// NewFault validates and sets the provided options and returns a Fault.
func NewFault(i Injector, opts ...Option) (*Fault, error) {
	// set the defaults.
	fault := &Fault{
		injector: i,
		randSeed: defaultRandSeed,
	}

	// apply the list of options to fault.
	for _, opt := range opts {
		opt.applyFault(fault)
	}

	// sanitize the options.
	if fault.injector == nil {
		return nil, ErrNilInjector
	}
	if fault.participation < 0 || fault.participation > 1.0 {
		return nil, ErrInvalidPercent
	}

	// set our random source with the provided seed.
	fault.rand = rand.New(rand.NewSource(fault.randSeed))

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
		} else {
			r = updateRequestContextValue(r, ContextValueDisabled)
		}

		// run the injector if shouldEvaluate and we're randomly chosen.
		if shouldEvaluate && f.percentDo() {
			f.injector.Handler(next).ServeHTTP(w, updateRequestContextValue(r, ContextValueInjected))
		} else {
			next.ServeHTTP(w, updateRequestContextValue(r, ContextValueSkipped))
		}
	})
}

// percentDo randomly decides (returns true) if the injector should run based on f.participation.
// Numbers outside of [0.0,1.0] will always return false.
func (f *Fault) percentDo() bool {
	rn := f.rand.Float32()
	if rn < f.participation && f.participation <= 1.0 {
		return true
	}

	return false
}
