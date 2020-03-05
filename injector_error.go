package fault

import (
	"errors"
	"net/http"
	"reflect"
)

var (
	// ErrInvalidHTTPCode when an invalid status code is provided.
	ErrInvalidHTTPCode = errors.New("not a valid http status code")
)

// ErrorInjector responds with an http status code and message.
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

// WithStatusText sets custom status text to write.
func WithStatusText(t string) ErrorInjectorOption {
	return statusTextOption(t)
}

func (o reporterOption) applyErrorInjector(i *ErrorInjector) error {
	i.reporter = o.reporter
	return nil
}

// NewErrorInjector returns an ErrorInjector that reponds with a status code.
func NewErrorInjector(code int, opts ...ErrorInjectorOption) (*ErrorInjector, error) {
	const placeholderStatusText = "go-fault: replace with default code text"

	// set defaults
	ei := &ErrorInjector{
		statusCode: code,
		statusText: placeholderStatusText,
		reporter:   NewNoopReporter(),
	}

	// apply options
	for _, opt := range opts {
		err := opt.applyErrorInjector(ei)
		if err != nil {
			return nil, err
		}
	}

	// check options
	if http.StatusText(ei.statusCode) == "" {
		return nil, ErrInvalidHTTPCode
	}
	if ei.statusText == placeholderStatusText {
		ei.statusText = http.StatusText(ei.statusCode)
	}

	return ei, nil
}

// Handler responds with the configured status code and text.
func (i *ErrorInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go i.reporter.Report(reflect.ValueOf(*i).Type().Name(), StateStarted)
		http.Error(w, i.statusText, i.statusCode)
		go i.reporter.Report(reflect.ValueOf(*i).Type().Name(), StateFinished)
	})
}
