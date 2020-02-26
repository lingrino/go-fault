package fault

import (
	"errors"
	"math/rand"
	"net/http"
	"reflect"
)

const (
	// defaultRandSeed is used when a random seed is not set explicitly
	defaultRandSeed = 1
)

var (
	// ErrNilInjector returns when a nil Injector type is passed.
	ErrNilInjector = errors.New("injector cannot be nil")
	// ErrInvalidPercent returns when a provided percent is outside of the allowed bounds.
	ErrInvalidPercent = errors.New("percent must be 0.0 <= percent <= 1.0")
)

// Fault is the main struct and combines an Injector with configuration.
type Fault struct {
	// opt holds all of our user provided fault options
	opt Options

	// pathBlacklist is a dict representation of Options.PathBlacklist that is populated in
	// NewFault and used to make path lookups faster.
	pathBlacklist map[string]bool

	// pathWhitelist is a dict representation of Options.PathWhitelist that is populated in
	// NewFault and used to make path lookups faster.
	pathWhitelist map[string]bool

	// randSeed is used to seed our random number generator
	rand *rand.Rand
}

// Options holds configuration for a Fault.
type Options struct {
	// Enabled determines if the fault middleware should evaluate.
	Enabled bool

	// PercentOfRequests is the percent of requests that should have the fault injected. 0.0 <=
	// percent <= 1.0
	PercentOfRequests float32

	// Injector is the interface that returns the handler we will inject.
	Injector Injector

	// PathBlacklist is a list of paths for which faults will never be injected
	PathBlacklist []string

	// PathWhitelist is a list of paths for which faults will be evaluated. If PathWhitelist is
	// empty then faults will evaluate on all paths.
	PathWhitelist []string

	// Reporter is an interface that receives fault event data at Reporter.Report and can act on
	// that data. By default we do not report on events and Reporter is nil.
	Reporter Reporter

	// RandSeed is a number to seed our random gnerator with. Only applies to Fault randomness,
	// not randomness used by injectors.
	RandSeed int64
}

// NewFault validates the provided options and returns a Fault struct.
func NewFault(o Options) (*Fault, error) {
	output := &Fault{}

	if o.Injector == nil {
		return nil, ErrNilInjector
	}

	if o.PercentOfRequests < 0 || o.PercentOfRequests > 1.0 {
		return nil, ErrInvalidPercent
	}

	if len(o.PathBlacklist) > 0 {
		output.pathBlacklist = make(map[string]bool, len(o.PathBlacklist))
		for _, path := range o.PathBlacklist {
			output.pathBlacklist[path] = true
		}
	}

	if len(o.PathWhitelist) > 0 {
		output.pathWhitelist = make(map[string]bool, len(o.PathWhitelist))
		for _, path := range o.PathWhitelist {
			output.pathWhitelist[path] = true
		}
	}

	if o.Reporter == nil {
		o.Reporter = NewNoopReporter()
	}

	// We assume that 0 is unspecified
	if o.RandSeed != 0 {
		output.rand = rand.New(rand.NewSource(o.RandSeed))
	} else {
		output.rand = rand.New(rand.NewSource(defaultRandSeed))
	}

	output.opt = o

	return output, nil
}

// Handler returns the main fault handler, which runs Injector.Handler a percent of the time.
func (f *Fault) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var shouldEvaluate bool

		// By default faults should not evaluate. Here we go through conditions where faults
		// will evaluate, if everything is configured correctly

		// f.opt.Enabled is the first check, to prioritize speed when faults are disabled
		if f.opt.Enabled && f.opt.Injector != nil {
			// If path is in blacklist, do not evaluate
			if _, ok := f.pathBlacklist[r.URL.Path]; !ok {
				// If whitelist exists and path is not in it, do not evaluate
				if len(f.pathWhitelist) > 0 {
					// If path is in the whitelist, evaluate
					if _, ok := f.pathWhitelist[r.URL.Path]; ok {
						shouldEvaluate = true
					}
				} else {
					// If whitelist does not exist, evaluate
					shouldEvaluate = true
				}
			}
		} else {
			r = updateRequestContextValue(r, ContextValueDisabled)
		}

		if shouldEvaluate && f.percentDo() {
			f.opt.Injector.Handler(next).ServeHTTP(w, updateRequestContextValue(r, ContextValueInjected))
		} else {
			f.opt.Reporter.Report(reflect.ValueOf(*f).Type().Name(), StateSkipped)
			next.ServeHTTP(w, updateRequestContextValue(r, ContextValueSkipped))
		}
	})
}

// percentDo takes a percent (0.0 <= per <= 1.0) and randomly returns true that percent of the time.
// Numbers provided outside of [0.0,1.0] will always return false.
func (f *Fault) percentDo() bool {
	var proceed bool

	rn := f.rand.Float32()
	if rn < f.opt.PercentOfRequests && f.opt.PercentOfRequests <= 1.0 {
		return true
	}

	return proceed
}
