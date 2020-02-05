package fault

import (
	"errors"
	"math/rand"
	"net/http"
	"time"
)

// ContextString is the type that all of our context keys will be
type ContextString string

// ContextValue is the value defined by ContextKey. It is a list of ContextString
// that will be added by Injectors
type ContextValue []ContextString

const (
	// ContextKey will be added to the request context of any injetor that does not return immediately
	// (ex: SlowInjector) and the value will be ContextValue, a list of ContextString that describe what
	// fault occurred.
	ContextKey ContextString = "fault-injected"
	// ContextValueError is added to ContextValue when an error (ex: misconfiguration)
	// occurred while trying to inject a fault
	ContextValueError ContextString = "fault-error"
	// ContextValueDisabled is added to ContextValue when the fault is disabled
	ContextValueDisabled ContextString = "fault-disabled"
	// ContextValueSlowInjector is added to ContextValue when the SlowInjector is injected
	ContextValueSlowInjector ContextString = "slow-injector"
)

var (
	// ErrNilInjector returns when a nil Injector type is passed.
	ErrNilInjector = errors.New("injector cannot be nil")
	// ErrInvalidPercent returns when a provided percent is outside of the allowed bounds.
	ErrInvalidPercent = errors.New("percent must be 0.0 <= percent <= 1.0")
	// ErrInvalidHTTPCode returns when an invalid http status code is provided.
	ErrInvalidHTTPCode = errors.New("not a valid http status code")
)

// Fault is the main struct and combines an Injector with configuration.
type Fault struct {
	opt Options

	// pathBlacklist is a dict representation of Options.PathBlacklist that is populated in
	// NewFault and used to make path lookups faster.
	pathBlacklist map[string]bool

	// pathWhitelist is a dict representation of Options.PathWhitelist that is populated in
	// NewFault and used to make path lookups faster.
	pathWhitelist map[string]bool
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

	output.opt = o

	// Seed rand for f.percentDo
	rand.Seed(time.Now().UnixNano())

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
		}

		if shouldEvaluate && f.percentDo() {
			f.opt.Injector.Handler(next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// percentDo takes a percent (0.0 <= per <= 1.0) and randomly returns true that percent of the time.
// Numbers provided outside of [0.0,1.0] will always return false.
func (f *Fault) percentDo() bool {
	var proceed bool

	rn := rand.Float32()
	if rn < f.opt.PercentOfRequests && f.opt.PercentOfRequests <= 1.0 {
		return true
	}

	return proceed
}

// Injector is an interface for our fault injection middleware. Injectors are wrapped into Faults.
// Faults handle running the Injector the correct percent of the time.
type Injector interface {
	Handler(next http.Handler) http.Handler
}

// ChainInjector combines many injectors into a single chain injector. In a chain injector the
// Handler func will execute ChainInjector.middlewares in order and then returns.
type ChainInjector struct {
	middlewares []func(next http.Handler) http.Handler
}

// NewChainInjector combines many injectors into a single chain injector. In a chain injector the
// Handler() for each injector will execute in the order provided.
func NewChainInjector(is ...Injector) (*ChainInjector, error) {
	chainInjector := &ChainInjector{}
	for _, i := range is {
		chainInjector.middlewares = append(chainInjector.middlewares, i.Handler)
	}

	return chainInjector, nil
}

// Handler executes ChainInjector.middlewares in order and then returns.
func (i *ChainInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil {
			// Loop in reverse to preserve handler order
			for idx := len(i.middlewares) - 1; idx >= 0; idx-- {
				next = i.middlewares[idx](next)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RandomInjector combines many injectors into a single injector. When the random injector is called
// it randomly runs one of the provided injectors.
type RandomInjector struct {
	randF       func(int) int
	middlewares []func(next http.Handler) http.Handler
}

// NewRandomInjector combines many injectors into a single random injector. When the random injector
// is called it randomly runs one of the provided injectors.
func NewRandomInjector(is ...Injector) (*RandomInjector, error) {
	rand.Seed(time.Now().UnixNano())
	RandomInjector := &RandomInjector{randF: rand.Intn}

	for _, i := range is {
		RandomInjector.middlewares = append(RandomInjector.middlewares, i.Handler)
	}

	return RandomInjector, nil
}

// Handler executes a random injector from RandomInjector.middlewares
func (i *RandomInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil && len(i.middlewares) > 0 {
			i.middlewares[i.randF(len(i.middlewares))](next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// RejectInjector immediately sends back an empty response.
type RejectInjector struct{}

// NewRejectInjector returns a RejectInjector struct.
func NewRejectInjector() (*RejectInjector, error) {
	return &RejectInjector{}, nil
}

// Handler immediately rejects the request, returning an empty response.
func (i *RejectInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a specialized and documented way of sending an interrupted response to
		// the client without printing the panic stack trace or erroring.
		// https://golang.org/pkg/net/http/#Handler
		panic(http.ErrAbortHandler)
	})
}

// ErrorInjector immediately responds with an http status code and the error message associated with
// that code.
type ErrorInjector struct {
	statusCode int
	statusText string
}

// NewErrorInjector returns an ErrorInjector that reponds with the configured status code.
func NewErrorInjector(code int) (*ErrorInjector, error) {
	statusText := http.StatusText(code)
	if statusText == "" {
		return nil, ErrInvalidHTTPCode
	}

	return &ErrorInjector{
		statusCode: code,
		statusText: statusText,
	}, nil
}

// Handler immediately responds with the configured HTTP status code and default status text for
// that code.
func (i *ErrorInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil {
			if http.StatusText(i.statusCode) != "" {
				http.Error(w, i.statusText, i.statusCode)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// SlowInjector sleeps a specified duration and then continues the request. Simulates latency.
type SlowInjector struct {
	duration time.Duration
	sleep    func(t time.Duration)
}

// NewSlowInjector returns a SlowInjector that adds the configured latency.
func NewSlowInjector(d time.Duration) (*SlowInjector, error) {
	return &SlowInjector{
		duration: d,
		sleep:    time.Sleep,
	}, nil
}

// Handler waits the configured duration and then continues the request.
func (i *SlowInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil {
			if i.sleep != nil {
				i.sleep(i.duration)
			}
		}
		next.ServeHTTP(w, updateRequestContextValue(r, ContextValueSlowInjector))
	})
}
