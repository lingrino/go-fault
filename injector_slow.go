package fault

import (
	"net/http"
	"time"
)

// SlowInjector runs slowF (default time.Sleep) and then continues the request. Simulates latency.
type SlowInjector struct {
	duration time.Duration
	slowF    func(t time.Duration)
	reporter Reporter
}

// SlowInjectorOption configures a SlowInjector.
type SlowInjectorOption interface {
	applySlowInjector(i *SlowInjector) error
}

type slowFunctionOption func(t time.Duration)

func (o slowFunctionOption) applySlowInjector(i *SlowInjector) error {
	i.slowF = o
	return nil
}

// WithSlowFunc sets the function that will be used to wait the time.Duration
func WithSlowFunc(f func(t time.Duration)) SlowInjectorOption {
	return slowFunctionOption(f)
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
		slowF:    time.Sleep,
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
		i.slowF(i.duration)
		next.ServeHTTP(w, r)
	})
}
