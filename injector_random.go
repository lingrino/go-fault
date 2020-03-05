package fault

import (
	"math/rand"
	"net/http"
	"sync"
)

// RandomInjector combines many injectors into a single injector. When the random injector is called
// it randomly runs one of the provided injectors.
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
// Make sure your function always returns an integer between [0,n) to avoid panics.
func WithRandIntFunc(f func(int) int) RandomInjectorOption {
	return randIntFuncOption(f)
}

// NewRandomInjector combines many injectors into a single random injector. When the random injector
// is called it randomly runs one of the provided injectors.
func NewRandomInjector(is []Injector, opts ...RandomInjectorOption) (*RandomInjector, error) {
	randomInjector := &RandomInjector{
		randSeed: defaultRandSeed,
		randF:    nil,
	}

	for _, opt := range opts {
		err := opt.applyRandomInjector(randomInjector)
		if err != nil {
			return nil, err
		}
	}

	for _, i := range is {
		randomInjector.middlewares = append(randomInjector.middlewares, i.Handler)
	}

	randomInjector.rand = rand.New(rand.NewSource(randomInjector.randSeed))
	if randomInjector.randF == nil {
		randomInjector.randF = randomInjector.rand.Intn
	}

	return randomInjector, nil
}

// Handler executes a random injector from RandomInjector.middlewares
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
