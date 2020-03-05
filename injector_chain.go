package fault

import "net/http"

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
