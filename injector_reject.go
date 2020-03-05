package fault

import (
	"net/http"
	"reflect"
)

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
		i.reporter.Report(reflect.ValueOf(*i).Type().Name(), StateStarted)

		// This is a specialized and documented way of sending an interrupted response to
		// the client without printing the panic stack trace or erroring.
		// https://golang.org/pkg/net/http/#Handler
		panic(http.ErrAbortHandler)
	})
}
