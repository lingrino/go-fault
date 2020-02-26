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

	// injectPercent is the percent of requests that run the injector. 0.0 <= p <= 1.0.
	injectPercent float32

	// pathBlacklist is a map of paths that the injector will not run against.
	pathBlacklist map[string]bool

	// pathWhitelist, if set, is a map of the only paths that the injector will run against.
	pathWhitelist map[string]bool

	// randSeed is a number to seed rand with.
	randSeed int64

	// rand is our random number source.
	rand *rand.Rand
}

// FaultOption configures a Fault.
type FaultOption interface {
	applyFault(s *Fault)
}

type enabledFaultOption bool

func (e enabledFaultOption) applyFault(f *Fault) {
	f.enabled = bool(e)
}

// WithEnabled determines if the fault should evaluate.
func WithEnabled(e bool) FaultOption {
	return enabledFaultOption(e)
}

type injectPercentFaultOption float32

func (p injectPercentFaultOption) applyFault(f *Fault) {
	f.injectPercent = float32(p)
}

// WithInjectPercent sets the percent of requests that run the injector. 0.0 <= p <= 1.0.
func WithInjectPercent(p float32) FaultOption {
	return injectPercentFaultOption(p)
}

type pathBlacklistFaultOption []string

func (p pathBlacklistFaultOption) applyFault(f *Fault) {
	blacklist := make(map[string]bool, len(p))
	for _, path := range p {
		blacklist[path] = true
	}
	f.pathBlacklist = blacklist
}

// WithPathBlacklist is a list of paths that the injector will not run against.
func WithPathBlacklist(blacklist []string) FaultOption {
	return pathBlacklistFaultOption(blacklist)
}

type pathWhitelistFaultOption []string

func (p pathWhitelistFaultOption) applyFault(f *Fault) {
	whitelist := make(map[string]bool, len(p))
	for _, path := range p {
		whitelist[path] = true
	}
	f.pathWhitelist = whitelist
}

// WithPathWhitelist is, if set, a map of the only paths that the injector will run against.
func WithPathWhitelist(whitelist []string) FaultOption {
	return pathWhitelistFaultOption(whitelist)
}

type randSeedFaultOption int64

func (p randSeedFaultOption) applyFault(f *Fault) {
	f.randSeed = int64(p)
}

// WithRandSeed sets the seed for fault.rand
func WithRandSeed(s int64) FaultOption {
	return randSeedFaultOption(s)
}

// NewFault validates and sets the provided options and returns a Fault.
func NewFault(i Injector, opts ...FaultOption) (*Fault, error) {
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
	if fault.injectPercent < 0 || fault.injectPercent > 1.0 {
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

// percentDo randomly decides (returns true) if the injector should run based on f.injectPercent.
// Numbers outside of [0.0,1.0] will always return false.
func (f *Fault) percentDo() bool {
	rn := f.rand.Float32()
	if rn < f.injectPercent && f.injectPercent <= 1.0 {
		return true
	}

	return false
}
