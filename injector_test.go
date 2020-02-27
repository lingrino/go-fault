package fault

import (
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewChainInjector tests NewChainInjector.
func TestNewChainInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    []Injector
		wantErr error
	}{
		{
			name:    "nil",
			give:    nil,
			wantErr: nil,
		},
		{
			name:    "empty",
			give:    []Injector{},
			wantErr: nil,
		},
		{
			name: "one",
			give: []Injector{
				newTestInjectorNoop(),
			},
			wantErr: nil,
		},
		{
			name: "two",
			give: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewChainInjector(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, len(tt.give), len(i.middlewares))
		})
	}
}

// TestChainInjectorHandler tests ChainInjector.Handler.
func TestChainInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     []Injector
		wantCode int
		wantBody string
	}{
		{
			name:     "nil",
			give:     nil,
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:     "empty",
			give:     []Injector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "one",
			give: []Injector{
				newTestInjectorOneOK(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "noop one",
			give: []Injector{
				newTestInjectorNoop(),
				newTestInjectorOneOK(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "two error",
			give: []Injector{
				newTestInjectorTwoTeapot(),
				newTestInjector500s(),
			},
			wantCode: http.StatusTeapot,
			wantBody: "two" + http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "one two",
			give: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorTwoTeapot(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + "two" + testHandlerBody,
		},
		{
			name: "one stop two",
			give: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorStop(),
				newTestInjectorTwoTeapot(),
			},
			wantCode: http.StatusOK,
			wantBody: "one",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ci, err := NewChainInjector(tt.give)
			assert.NoError(t, err)

			f, err := NewFault(ci,
				WithEnabled(true),
				WithInjectPercent(1.0),
			)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestNewRandomInjector tests NewRandomInjector.
func TestNewRandomInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveInjector []Injector
		giveOptions  []RandomInjectorOption
		wantRand     *rand.Rand
		wantErr      error
	}{
		{
			name:         "nil",
			giveInjector: nil,
			giveOptions:  nil,
			wantRand:     rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:      nil,
		},
		{
			name:         "empty",
			giveInjector: []Injector{},
			giveOptions:  nil,
			wantRand:     rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:      nil,
		},
		{
			name: "one",
			giveInjector: []Injector{
				newTestInjectorNoop(),
			},
			giveOptions: nil,
			wantRand:    rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:     nil,
		},
		{
			name: "two",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			giveOptions: nil,
			wantRand:    rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:     nil,
		},
		{
			name: "with seed",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			giveOptions: []RandomInjectorOption{
				WithRandSeed(100),
			},
			wantRand: rand.New(rand.NewSource(100)),
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewRandomInjector(tt.giveInjector, tt.giveOptions...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, len(tt.giveInjector), len(i.middlewares))
		})
	}
}

// // TestRandomInjectorHandler tests RandomInjector.Handler.
// func TestRandomInjectorHandler(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name     string
// 		give     *RandomInjector
// 		wantCode int
// 		wantBody string
// 	}{
// 		{
// 			name:     "nil",
// 			give:     nil,
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name:     "empty",
// 			give:     &RandomInjector{},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "nil middlewares",
// 			give: &RandomInjector{
// 				middlewares: nil,
// 			},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "empty middlewares",
// 			give: &RandomInjector{
// 				middlewares: []func(next http.Handler) http.Handler{},
// 			},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "one",
// 			give: &RandomInjector{
// 				randF: func(int) int { return 0 },
// 				middlewares: []func(next http.Handler) http.Handler{
// 					func(next http.Handler) http.Handler {
// 						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 							w.WriteHeader(http.StatusOK)
// 							fmt.Fprint(w, "one")
// 							next.ServeHTTP(w, r)
// 						})
// 					},
// 				},
// 			},
// 			wantCode: http.StatusOK,
// 			wantBody: "one" + testHandlerBody,
// 		},
// 		{
// 			name: "two",
// 			give: &RandomInjector{
// 				randF: func(int) int { return 1 },
// 				middlewares: []func(next http.Handler) http.Handler{
// 					func(next http.Handler) http.Handler {
// 						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 							w.WriteHeader(http.StatusOK)
// 							fmt.Fprint(w, "one")
// 							next.ServeHTTP(w, r)
// 						})
// 					},
// 					func(next http.Handler) http.Handler {
// 						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 							w.WriteHeader(http.StatusOK)
// 							fmt.Fprint(w, "two")
// 							next.ServeHTTP(w, r)
// 						})
// 					},
// 				},
// 			},
// 			wantCode: http.StatusOK,
// 			wantBody: "two" + testHandlerBody,
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			f := &Fault{
// 				opt: Options{
// 					Enabled:           true,
// 					Injector:          tt.give,
// 					PercentOfRequests: 1.0,
// 				},
// 				rand: rand.New(rand.NewSource(defaultRandSeed)),
// 			}

// 			rr := testRequest(t, f)

// 			assert.Equal(t, tt.wantCode, rr.Code)
// 			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
// 		})
// 	}
// }

// // TestNewErrorInjector tests NewErrorInjector.
// func TestNewErrorInjector(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		give    int
// 		want    *ErrorInjector
// 		wantErr error
// 	}{
// 		{
// 			give:    -1,
// 			want:    nil,
// 			wantErr: ErrInvalidHTTPCode,
// 		},
// 		{
// 			give:    0,
// 			want:    nil,
// 			wantErr: ErrInvalidHTTPCode,
// 		},
// 		{
// 			give: testHandlerCode,
// 			want: &ErrorInjector{
// 				statusCode: testHandlerCode,
// 				statusText: testHandlerBody,
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			give: http.StatusInternalServerError,
// 			want: &ErrorInjector{
// 				statusCode: http.StatusInternalServerError,
// 				statusText: http.StatusText(http.StatusInternalServerError),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			give:    120000,
// 			want:    nil,
// 			wantErr: ErrInvalidHTTPCode,
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
// 			t.Parallel()

// 			i, err := NewErrorInjector(tt.give)

// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.want, i)
// 		})
// 	}
// }

// // TestNewRejectInjector tests NewRejectInjector.
// func TestNewRejectInjector(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name    string
// 		want    *RejectInjector
// 		wantErr error
// 	}{
// 		{
// 			name:    "valid",
// 			want:    &RejectInjector{},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			i, err := NewRejectInjector()

// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.want, i)
// 		})
// 	}
// }

// // TestRejectInjectorHandler tests RejectInjector.Handler.
// func TestRejectInjectorHandler(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name string
// 		give *RejectInjector
// 	}{
// 		{
// 			name: "valid nil",
// 			give: nil,
// 		},
// 		{
// 			name: "valid empty",
// 			give: &RejectInjector{},
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			f := &Fault{
// 				opt: Options{
// 					Enabled:           true,
// 					Injector:          tt.give,
// 					PercentOfRequests: 1.0,
// 				},
// 				rand: rand.New(rand.NewSource(defaultRandSeed)),
// 			}

// 			rr := testRequestExpectPanic(t, f)

// 			assert.Nil(t, rr)
// 		})
// 	}
// }

// // TestErrorInjectorHandler tests ErrorInjector.Handler.
// func TestErrorInjectorHandler(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name     string
// 		give     *ErrorInjector
// 		wantCode int
// 		wantBody string
// 	}{
// 		{
// 			name:     "nil",
// 			give:     nil,
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name:     "empty",
// 			give:     &ErrorInjector{},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "1",
// 			give: &ErrorInjector{
// 				statusCode: 1,
// 				statusText: "one",
// 			},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "200",
// 			give: &ErrorInjector{
// 				statusCode: testHandlerCode,
// 				statusText: testHandlerBody,
// 			},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "418",
// 			give: &ErrorInjector{
// 				statusCode: http.StatusTeapot,
// 				statusText: http.StatusText(http.StatusTeapot),
// 			},
// 			wantCode: http.StatusTeapot,
// 			wantBody: http.StatusText(http.StatusTeapot),
// 		},
// 		{
// 			name: "500",
// 			give: &ErrorInjector{
// 				statusCode: http.StatusInternalServerError,
// 				statusText: http.StatusText(http.StatusInternalServerError),
// 			},
// 			wantCode: http.StatusInternalServerError,
// 			wantBody: http.StatusText(http.StatusInternalServerError),
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			f := &Fault{
// 				opt: Options{
// 					Enabled:           true,
// 					Injector:          tt.give,
// 					PercentOfRequests: 1.0,
// 				},
// 				rand: rand.New(rand.NewSource(defaultRandSeed)),
// 			}

// 			rr := testRequest(t, f)

// 			assert.Equal(t, tt.wantCode, rr.Code)
// 			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
// 		})
// 	}
// }

// // TestNewSlowInjector tests NewSlowInjector.
// func TestNewSlowInjector(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		give    time.Duration
// 		want    *SlowInjector
// 		wantErr error
// 	}{
// 		{
// 			give: 0,
// 			want: &SlowInjector{
// 				duration: 0,
// 				sleep:    time.Sleep,
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			give: time.Millisecond,
// 			want: &SlowInjector{
// 				duration: time.Millisecond,
// 				sleep:    time.Sleep,
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			give: time.Millisecond * 1000,
// 			want: &SlowInjector{
// 				duration: time.Second,
// 				sleep:    time.Sleep,
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			give: time.Hour * 1000000,
// 			want: &SlowInjector{
// 				duration: time.Hour * 1000000,
// 				sleep:    time.Sleep,
// 			},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
// 			t.Parallel()

// 			i, err := NewSlowInjector(tt.give)

// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.want.duration, i.duration)
// 		})
// 	}
// }

// // TestSlowInjectorHandler tests SlowInjector.Handler.
// func TestSlowInjectorHandler(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name     string
// 		give     *SlowInjector
// 		wantCode int
// 		wantBody string
// 	}{
// 		{
// 			name:     "nil",
// 			give:     nil,
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name:     "empty",
// 			give:     &SlowInjector{},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 		{
// 			name: "valid",
// 			give: &SlowInjector{
// 				duration: time.Millisecond,
// 				sleep:    func(d time.Duration) {},
// 			},
// 			wantCode: testHandlerCode,
// 			wantBody: testHandlerBody,
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			f := &Fault{
// 				opt: Options{
// 					Enabled:           true,
// 					Injector:          tt.give,
// 					PercentOfRequests: 1.0,
// 				},
// 				rand: rand.New(rand.NewSource(defaultRandSeed)),
// 			}

// 			rr := testRequest(t, f)

// 			assert.Equal(t, tt.wantCode, rr.Code)
// 			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
// 		})
// 	}
// }
