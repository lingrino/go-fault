package fault

import (
	"net/http"
	"reflect"
	"time"
)

// SlowInjector waits and then continues the request.
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

// WithSlowFunc sets the function that will be used to wait the time.Duration.
func WithSlowFunc(f func(t time.Duration)) SlowInjectorOption {
	return slowFunctionOption(f)
}

func (o reporterOption) applySlowInjector(i *SlowInjector) error {
	i.reporter = o.reporter
	return nil
}

// NewSlowInjector returns a SlowInjector.
func NewSlowInjector(d time.Duration, opts ...SlowInjectorOption) (*SlowInjector, error) {
	// set defaults
	si := &SlowInjector{
		duration: d,
		slowF:    time.Sleep,
		reporter: NewNoopReporter(),
	}

	// apply options
	for _, opt := range opts {
		err := opt.applySlowInjector(si)
		if err != nil {
			return nil, err
		}
	}

	return si, nil
}

// Handler runs i.slowF to wait the set duration and then continues.
func (i *SlowInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go i.reporter.Report(reflect.ValueOf(*i).Type().Name(), StateStarted)
		i.slowF(i.duration)
		go i.reporter.Report(reflect.ValueOf(*i).Type().Name(), StateFinished)

		next.ServeHTTP(w, r)
	})
}
