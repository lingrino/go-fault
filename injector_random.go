package fault

import (
	"math/rand"
	"net/http"
	"sync"
)

// RandomInjector combines many Injectors into a single Injector that runs one randomly.
type RandomInjector struct {
	middlewares []func(next http.Handler) http.Handler

	randSeed int64
	rand     *rand.Rand
	randF    func(int) int

	// *rand.Rand is not thread safe. This mutex protects our random source
	randMtx sync.Mutex
}

// RandomInjectorOption configures a RandomInjector.
type RandomInjectorOption interface {
	applyRandomInjector(i *RandomInjector) error
}

func (o randSeedOption) applyRandomInjector(i *RandomInjector) error {
	i.randSeed = int64(o)
	return nil
}

type randIntFuncOption func(int) int

func (o randIntFuncOption) applyRandomInjector(i *RandomInjector) error {
	i.randF = o
	return nil
}

// WithRandIntFunc sets the function that will be used to randomly get an int. Default rand.Intn.
// Always returns an integer between [0,n) to avoid panics.
func WithRandIntFunc(f func(int) int) RandomInjectorOption {
	return randIntFuncOption(f)
}

// NewRandomInjector combines many Injectors into a single Injector that runs one randomly.
func NewRandomInjector(is []Injector, opts ...RandomInjectorOption) (*RandomInjector, error) {
	// set defaults
	ri := &RandomInjector{
		randSeed: defaultRandSeed,
		randF:    nil,
	}

	// apply options
	for _, opt := range opts {
		err := opt.applyRandomInjector(ri)
		if err != nil {
			return nil, err
		}
	}

	// set middleware
	for _, i := range is {
		ri.middlewares = append(ri.middlewares, i.Handler)
	}

	// set seeded rand source and function
	ri.rand = rand.New(rand.NewSource(ri.randSeed))
	if ri.randF == nil {
		ri.randF = ri.rand.Intn
	}

	return ri, nil
}

// Handler executes a random Injector from RandomInjector.middlewares.
func (i *RandomInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(i.middlewares) > 0 {
			i.randMtx.Lock()
			randIdx := i.randF(len(i.middlewares))
			i.randMtx.Unlock()

			i.middlewares[randIdx](next).ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
