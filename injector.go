package fault

import (
	"errors"
	"math/rand"
	"net/http"
	"reflect"
	"time"
)

// InjectorState represents the states an injector can be in
type InjectorState int

const (
	StateStarted InjectorState = iota + 1
	StateFinished
	StateSkipped
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
	applyRandomInjector(i *RandomInjector) error
}

func (o randSeedOption) applyRandomInjector(i *RandomInjector) error {
	i.randSeed = int64(o)
	return nil
}

// NewRandomInjector combines many injectors into a single random injector. When the random injector
// is called it randomly runs one of the provided injectors.
func NewRandomInjector(is []Injector, opts ...RandomInjectorOption) (*RandomInjector, error) {
	randomInjector := &RandomInjector{
		randSeed: defaultRandSeed,
	}

	for _, opt := range opts {
		err := opt.applyRandomInjector(randomInjector)
		if err != nil {
			return nil, err
		}
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
			i.middlewares[i.rand.Intn(len(i.middlewares))](next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// RejectInjector immediately sends back an empty response.
type RejectInjector struct {
	reporter Reporter
}

// RejectInjectorOption configures a RejectInjector.
type RejectInjectorOption interface {
	applyRejectInjector(i *RejectInjector) error
}

func (o reporterOption) applyRejectInjector(i *RejectInjector) error {
	i.reporter = o.reporter
	return nil
}

// NewRejectInjector returns a RejectInjector struct.
func NewRejectInjector(opts ...RejectInjectorOption) (*RejectInjector, error) {
	// set the defaults.
	ri := &RejectInjector{
		reporter: NewNoopReporter(),
	}

	// apply the options.
	for _, opt := range opts {
		err := opt.applyRejectInjector(ri)
		if err != nil {
			return nil, err
		}
	}

	return ri, nil
}

// Handler immediately rejects the request, returning an empty response.
func (i *RejectInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil {
			i.reporter.Report(reflect.ValueOf(*i).Type().Name(), StateStarted)
		}

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
	reporter   Reporter
}

// ErrorInjectorOption configures an ErrorInjector.
type ErrorInjectorOption interface {
	applyErrorInjector(i *ErrorInjector) error
}

type statusTextOption string

func (o statusTextOption) applyErrorInjector(i *ErrorInjector) error {
	i.statusText = string(o)
	return nil
}

// WithStatusText sets the status text that should return.
func WithStatusText(t string) ErrorInjectorOption {
	return statusTextOption(t)
}

func (o reporterOption) applyErrorInjector(i *ErrorInjector) error {
	i.reporter = o.reporter
	return nil
}

// NewErrorInjector returns an ErrorInjector that reponds with the configured status code.
func NewErrorInjector(code int, opts ...ErrorInjectorOption) (*ErrorInjector, error) {
	const placeholderStatusText = "go-fault replace with default code text"

	// set the defaults. by default we return ErrInvalidHTTPCode since 0 is invalid.
	ei := &ErrorInjector{
		statusCode: code,
		statusText: placeholderStatusText,
		reporter:   NewNoopReporter(),
	}

	// apply the options.
	for _, opt := range opts {
		err := opt.applyErrorInjector(ei)
		if err != nil {
			return nil, err
		}
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
	})
}

// SlowInjector sleeps a specified duration and then continues the request. Simulates latency.
type SlowInjector struct {
	duration time.Duration
	sleep    func(t time.Duration)
	reporter Reporter
}

// SlowInjectorOption configures a SlowInjector.
type SlowInjectorOption interface {
	applySlowInjector(i *SlowInjector) error
}

type sleepFunctionOption func(t time.Duration)

func (o sleepFunctionOption) applySlowInjector(i *SlowInjector) error {
	i.sleep = o
	return nil
}

// WithSleepFunction sets the function that will be used to wait the time.Duration
func WithSleepFunction(f func(t time.Duration)) SlowInjectorOption {
	return sleepFunctionOption(f)
}

func (o reporterOption) applySlowInjector(i *SlowInjector) error {
	i.reporter = o.reporter
	return nil
}

// NewSlowInjector returns a SlowInjector that adds the configured latency.
func NewSlowInjector(d time.Duration, opts ...SlowInjectorOption) (*SlowInjector, error) {
	// set the defaults.
	si := &SlowInjector{
		duration: d,
		sleep:    time.Sleep,
		reporter: NewNoopReporter(),
	}

	// apply the options.
	for _, opt := range opts {
		err := opt.applySlowInjector(si)
		if err != nil {
			return nil, err
		}
	}

	return si, nil
}

// Handler waits the configured duration and then continues the request.
func (i *SlowInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i.sleep(i.duration)
		next.ServeHTTP(w, r)
	})
}
