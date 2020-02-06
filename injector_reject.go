package fault

import "net/http"

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
