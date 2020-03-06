package fault

import "net/http"

// ChainInjector combines many Injectors into a single Injector that runs them in order.
type ChainInjector struct {
	middlewares []func(next http.Handler) http.Handler
}

// ChainInjectorOption configures a ChainInjector.
type ChainInjectorOption interface {
	applyChainInjector(i *ChainInjector) error
}

// NewChainInjector combines many Injectors into a single Injector that runs them in order.
func NewChainInjector(is []Injector, opts ...ChainInjectorOption) (*ChainInjector, error) {
	// set defaults
	ci := &ChainInjector{}

	// apply options
	for _, opt := range opts {
		err := opt.applyChainInjector(ci)
		if err != nil {
			return nil, err
		}
	}

	// set middleware
	for _, i := range is {
		ci.middlewares = append(ci.middlewares, i.Handler)
	}

	return ci, nil
}

// Handler executes ChainInjector.middlewares in order and then returns.
func (i *ChainInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Loop in reverse to preserve handler order
		for idx := len(i.middlewares) - 1; idx >= 0; idx-- {
			next = i.middlewares[idx](next)
		}

		next.ServeHTTP(w, r)
	})
}
