package fault

import (
	"math/rand"
	"net/http"
)

// RandomInjector combines many injectors into a single injector. When the random injector is called
// it randomly runs one of the provided injectors.
type RandomInjector struct {
	randF       func(int) int
	middlewares []func(next http.Handler) http.Handler
}

// NewRandomInjector combines many injectors into a single random injector. When the random injector
// is called it randomly runs one of the provided injectors.
func NewRandomInjector(is ...Injector) (*RandomInjector, error) {
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
