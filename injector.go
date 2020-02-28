package fault

import (
	"errors"
	"math/rand"
	"net/http"
	"time"
)

var (
	// ErrInvalidHTTPCode returns when an invalid http status code is provided.
	ErrInvalidHTTPCode = errors.New("not a valid http status code")
)

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
func NewChainInjector(is []Injector) (*ChainInjector, error) {
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
			r = updateRequestContextValue(r, ContextValueChainInjector)

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
	middlewares []func(next http.Handler) http.Handler

	randSeed int64
	rand     *rand.Rand
}

// RandomInjectorOption configures a RandomInjector.
type RandomInjectorOption interface {
	applyErrorInjector(i *RandomInjector)
}

func (o randSeedOption) applyErrorInjector(i *RandomInjector) {
	i.randSeed = int64(o)
}

// NewRandomInjector combines many injectors into a single random injector. When the random injector
// is called it randomly runs one of the provided injectors.
func NewRandomInjector(is []Injector, opts ...RandomInjectorOption) (*RandomInjector, error) {
	randomInjector := &RandomInjector{
		randSeed: defaultRandSeed,
	}

	for _, opt := range opts {
		opt.applyErrorInjector(randomInjector)
	}

	for _, i := range is {
		randomInjector.middlewares = append(randomInjector.middlewares, i.Handler)
	}

	randomInjector.rand = rand.New(rand.NewSource(randomInjector.randSeed))

	return randomInjector, nil
}

// Handler executes a random injector from RandomInjector.middlewares
func (i *RandomInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil && len(i.middlewares) > 0 {
			r = updateRequestContextValue(r, ContextValueRandomInjector)
			i.middlewares[i.rand.Intn(len(i.middlewares))](next).ServeHTTP(w, r)
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

// ErrorInjectorOption configures an ErrorInjector.
type ErrorInjectorOption interface {
	applyErrorInjector(i *ErrorInjector)
}

type statusTextOption string

func (o statusTextOption) applyErrorInjector(i *ErrorInjector) {
	i.statusText = string(o)
}

// WithStatusText sets the status text that should return.
func WithStatusText(t string) ErrorInjectorOption {
	return statusTextOption(t)
}

// NewErrorInjector returns an ErrorInjector that reponds with the configured status code.
func NewErrorInjector(code int, opts ...ErrorInjectorOption) (*ErrorInjector, error) {
	const placeholderStatusText = "go-fault replace with default code text"

	// set the defaults.
	// by default we return ErrInvalidHTTPCode since 0 is invalid.
	ei := &ErrorInjector{
		statusCode: code,
		statusText: placeholderStatusText,
	}

	// apply the options.
	for _, opt := range opts {
		opt.applyErrorInjector(ei)
	}

	// sanitize the options.
	if http.StatusText(ei.statusCode) == "" {
		return nil, ErrInvalidHTTPCode
	}
	if ei.statusText == placeholderStatusText {
		ei.statusText = http.StatusText(ei.statusCode)
	}

	return ei, nil
}

// Handler immediately responds with the configured HTTP status code text.
func (i *ErrorInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, i.statusText, i.statusCode)
		return
	})
}

// SlowInjector sleeps a specified duration and then continues the request. Simulates latency.
type SlowInjector struct {
	duration time.Duration
	sleep    func(t time.Duration)
}

// SlowInjectorOption configures an SlowInjector.
type SlowInjectorOption interface {
	applySlowInjector(i *SlowInjector)
}

type sleepFunctionOption func(t time.Duration)

func (o sleepFunctionOption) applySlowInjector(i *SlowInjector) {
	i.sleep = o
}

// WithSleepFunction sets the function that will be used to wait the time.Duration
func WithSleepFunction(f func(t time.Duration)) SlowInjectorOption {
	return sleepFunctionOption(f)
}

// NewSlowInjector returns a SlowInjector that adds the configured latency.
func NewSlowInjector(d time.Duration, opts ...SlowInjectorOption) (*SlowInjector, error) {
	// set the defaults.
	si := &SlowInjector{
		duration: d,
		sleep:    time.Sleep,
	}

	// apply the options.
	for _, opt := range opts {
		opt.applySlowInjector(si)
	}

	return si, nil
}

// Handler waits the configured duration and then continues the request.
func (i *SlowInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i.sleep(i.duration)
		next.ServeHTTP(w, updateRequestContextValue(r, ContextValueSlowInjector))
	})
}
