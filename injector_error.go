package fault

import "net/http"

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
