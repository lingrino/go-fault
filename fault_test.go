package fault

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// NOTE - Even though the New() constructors should prevent nil Faults/Injectors/etc...
//        we still want to test how the Faults and Injectors behave with unwanted input
//        and we expect that they handle it gracefully.

// TestNewFault tests NewFault.
func TestNewFault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		give      Options
		wantFault *Fault
		wantErr   error
	}{
		{
			name: "valid",
			give: Options{
				Enabled:           true,
				Injector:          newTestInjector(false),
				PercentOfRequests: 1.0,
				PathBlacklist: []string{
					"/donotinject",
				},
				PathWhitelist: []string{
					"/faultenabled",
				},
			},
			wantFault: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: false,
					},
					PercentOfRequests: 1.0,
					PathBlacklist: []string{
						"/donotinject",
					},
					PathWhitelist: []string{
						"/faultenabled",
					},
				},
				pathBlacklist: map[string]bool{
					"/donotinject": true,
				},
				pathWhitelist: map[string]bool{
					"/faultenabled": true,
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid injector",
			give: Options{
				Injector:          nil,
				PercentOfRequests: 1.0,
			},
			wantFault: nil,
			wantErr:   ErrNilInjector,
		},
		{
			name: "invalid percent",
			give: Options{
				Injector:          newTestInjector(false),
				PercentOfRequests: 1.1,
			},
			wantFault: nil,
			wantErr:   ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantFault, f)
		})
	}
}

// TestFaultHandler tests Fault.Handler.
func TestFaultHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *Fault
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
			give:     &Fault{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "nil injector",
			give: &Fault{
				opt: Options{
					Enabled:           true,
					Injector:          nil,
					PercentOfRequests: 1.0,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "not enabled",
			give: &Fault{
				opt: Options{
					Enabled: false,
					Injector: &testInjector{
						resp500: false,
					},
					PercentOfRequests: 1.0,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "zero percent",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: false,
					},
					PercentOfRequests: 0.0,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "100 percent 500s",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: true,
					},
					PercentOfRequests: 1.0,
				},
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "100 percent 500s with blacklist root",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: true,
					},
					PercentOfRequests: 1.0,
					PathBlacklist: []string{
						"/",
					},
				},
				pathBlacklist: map[string]bool{
					"/": true,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "100 percent 500s with whitelist root",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: true,
					},
					PercentOfRequests: 1.0,
					PathWhitelist: []string{
						"/",
					},
				},
				pathWhitelist: map[string]bool{
					"/": true,
				},
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "100 percent 500s with whitelist other",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: true,
					},
					PercentOfRequests: 1.0,
					PathWhitelist: []string{
						"/onlyinject",
					},
				},
				pathWhitelist: map[string]bool{
					"/onlyinject": true,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "100 percent 500s with whitelist and blacklist root",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: true,
					},
					PercentOfRequests: 1.0,
					PathBlacklist: []string{
						"/",
					},
					PathWhitelist: []string{
						"/",
					},
				},
				pathBlacklist: map[string]bool{
					"/": true,
				},
				pathWhitelist: map[string]bool{
					"/": true,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "100 percent",
			give: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: false,
					},
					PercentOfRequests: 1.0,
				},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rr := testRequest(t, tt.give)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestFaultPercentDo tests the internal Fault.percentDo.
func TestFaultPercentDo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		givePercent float32
		wantPercent float32
		wantRange   float32
	}{
		{-1.0, 0.0, 0.0},
		{},
		{0.0, 0.0, 0.0},
		{0.0001, 0.0001, 0.005},
		{0.3298, 0.3298, 0.005},
		{0.75, 0.75, 0.005},
		{1.0, 1.0, 0.0},
		{1.1, 0.0, 0.0},
		{10000.1, 0.0, 0.0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%g", tt.givePercent), func(t *testing.T) {
			t.Parallel()

			f := &Fault{
				opt: Options{
					PercentOfRequests: tt.givePercent,
				},
			}

			var errorC, totalC float32
			for totalC <= 100000 {
				result := f.percentDo()
				if result {
					errorC++
				}
				totalC++
			}

			minP := tt.wantPercent - tt.wantRange
			per := errorC / totalC
			maxP := tt.wantPercent + tt.wantRange

			assert.GreaterOrEqual(t, per, minP)
			assert.LessOrEqual(t, per, maxP)
		})
	}
}

// TestNewChainInjector tests NewChainInjector.
func TestNewChainInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    []Injector
		wantLen int
		wantNil bool
		wantErr error
	}{
		{
			name:    "nil",
			give:    nil,
			wantLen: 0,
			wantNil: false,
			wantErr: nil,
		},
		{
			name:    "empty",
			give:    []Injector{},
			wantLen: 0,
			wantNil: false,
			wantErr: nil,
		},
		{
			name: "one",
			give: []Injector{
				&SlowInjector{
					duration: time.Millisecond,
					sleep:    time.Sleep,
				},
			},
			wantLen: 1,
			wantNil: false,
			wantErr: nil,
		},
		{
			name: "two",
			give: []Injector{
				&SlowInjector{
					duration: time.Millisecond,
					sleep:    time.Sleep,
				},
				&ErrorInjector{
					statusCode: http.StatusInternalServerError,
					statusText: http.StatusText(http.StatusInternalServerError),
				},
			},
			wantLen: 2,
			wantNil: false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewChainInjector(tt.give...)

			assert.Equal(t, tt.wantErr, err)
			if tt.wantNil {
				assert.Nil(t, i)
			} else {
				assert.Equal(t, tt.wantLen, len(i.middlewares))
			}
		})
	}
}

// TestChainInjectorHandler tests ChainInjector.Handler.
func TestChainInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *ChainInjector
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
			give:     &ChainInjector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "nil middlewares",
			give: &ChainInjector{
				middlewares: nil,
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "empty middlewares",
			give: &ChainInjector{
				middlewares: []func(next http.Handler) http.Handler{},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "one continue",
			give: &ChainInjector{
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "one")
							next.ServeHTTP(w, r)
						})
					},
				},
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "one halt",
			give: &ChainInjector{
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusNoContent)
							fmt.Fprint(w, "one")
							return
						})
					},
				},
			},
			wantCode: http.StatusNoContent,
			wantBody: "one",
		},
		{
			name: "two continue",
			give: &ChainInjector{
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "one")
							next.ServeHTTP(w, r)
						})
					},
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "two")
							next.ServeHTTP(w, r)
						})
					},
				},
			},
			wantCode: http.StatusOK,
			wantBody: "onetwo" + testHandlerBody,
		},
		{
			name: "two halting",
			give: &ChainInjector{
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "one")
							return
						})
					},
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "two")
							return
						})
					},
				},
			},
			wantCode: http.StatusOK,
			wantBody: "one",
		},
		{
			name: "continue then halt",
			give: &ChainInjector{
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusTeapot)
							fmt.Fprint(w, "one")
							next.ServeHTTP(w, r)
						})
					},
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "two")
							return
						})
					},
				},
			},
			wantCode: http.StatusTeapot,
			wantBody: "onetwo",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &Fault{
				opt: Options{
					Enabled:           true,
					Injector:          tt.give,
					PercentOfRequests: 1.0,
				},
			}

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
		name    string
		give    []Injector
		wantLen int
		wantNil bool
		wantErr error
	}{
		{
			name:    "nil",
			give:    nil,
			wantLen: 0,
			wantNil: false,
			wantErr: nil,
		},
		{
			name:    "empty",
			give:    []Injector{},
			wantLen: 0,
			wantNil: false,
			wantErr: nil,
		},
		{
			name: "one",
			give: []Injector{
				&SlowInjector{
					duration: time.Millisecond,
					sleep:    time.Sleep,
				},
			},
			wantLen: 1,
			wantNil: false,
			wantErr: nil,
		},
		{
			name: "two",
			give: []Injector{
				&SlowInjector{
					duration: time.Millisecond,
					sleep:    time.Sleep,
				},
				&ErrorInjector{
					statusCode: http.StatusInternalServerError,
					statusText: http.StatusText(http.StatusInternalServerError),
				},
			},
			wantLen: 2,
			wantNil: false,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewRandomInjector(tt.give...)

			assert.Equal(t, tt.wantErr, err)
			if tt.wantNil {
				assert.Nil(t, i)
			} else {
				assert.Equal(t, tt.wantLen, len(i.middlewares))
			}
		})
	}
}

// TestRandomInjectorHandler tests RandomInjector.Handler.
func TestRandomInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *RandomInjector
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
			give:     &RandomInjector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "nil middlewares",
			give: &RandomInjector{
				middlewares: nil,
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "empty middlewares",
			give: &RandomInjector{
				middlewares: []func(next http.Handler) http.Handler{},
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "one",
			give: &RandomInjector{
				randF: func(int) int { return 0 },
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "one")
							next.ServeHTTP(w, r)
						})
					},
				},
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "two",
			give: &RandomInjector{
				randF: func(int) int { return 1 },
				middlewares: []func(next http.Handler) http.Handler{
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "one")
							next.ServeHTTP(w, r)
						})
					},
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "two")
							next.ServeHTTP(w, r)
						})
					},
				},
			},
			wantCode: http.StatusOK,
			wantBody: "two" + testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &Fault{
				opt: Options{
					Enabled:           true,
					Injector:          tt.give,
					PercentOfRequests: 1.0,
				},
			}

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestNewRejectInjector tests NewRejectInjector.
func TestNewRejectInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    *RejectInjector
		wantErr error
	}{
		{
			name:    "valid",
			want:    &RejectInjector{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewRejectInjector()

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, i)
		})
	}
}

// TestRejectInjectorHandler tests RejectInjector.Handler.
func TestRejectInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give *RejectInjector
	}{
		{
			name: "valid nil",
			give: nil,
		},
		{
			name: "valid empty",
			give: &RejectInjector{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &Fault{
				opt: Options{
					Enabled:           true,
					Injector:          tt.give,
					PercentOfRequests: 1.0,
				},
			}

			rr := testRequestExpectPanic(t, f)

			assert.Nil(t, rr)
		})
	}
}

// TestNewErrorInjector tests NewErrorInjector.
func TestNewErrorInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    int
		want    *ErrorInjector
		wantErr error
	}{
		{
			give:    -1,
			want:    nil,
			wantErr: ErrInvalidHTTPCode,
		},
		{
			give:    0,
			want:    nil,
			wantErr: ErrInvalidHTTPCode,
		},
		{
			give: testHandlerCode,
			want: &ErrorInjector{
				statusCode: testHandlerCode,
				statusText: testHandlerBody,
			},
			wantErr: nil,
		},
		{
			give: http.StatusInternalServerError,
			want: &ErrorInjector{
				statusCode: http.StatusInternalServerError,
				statusText: http.StatusText(http.StatusInternalServerError),
			},
			wantErr: nil,
		},
		{
			give:    120000,
			want:    nil,
			wantErr: ErrInvalidHTTPCode,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
			t.Parallel()

			i, err := NewErrorInjector(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, i)

		})
	}
}

// TestErrorInjectorHandler tests ErrorInjector.Handler.
func TestErrorInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *ErrorInjector
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
			give:     &ErrorInjector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "1",
			give: &ErrorInjector{
				statusCode: 1,
				statusText: "one",
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "200",
			give: &ErrorInjector{
				statusCode: testHandlerCode,
				statusText: testHandlerBody,
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "418",
			give: &ErrorInjector{
				statusCode: http.StatusTeapot,
				statusText: http.StatusText(http.StatusTeapot),
			},
			wantCode: http.StatusTeapot,
			wantBody: http.StatusText(http.StatusTeapot),
		},
		{
			name: "500",
			give: &ErrorInjector{
				statusCode: http.StatusInternalServerError,
				statusText: http.StatusText(http.StatusInternalServerError),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &Fault{
				opt: Options{
					Enabled:           true,
					Injector:          tt.give,
					PercentOfRequests: 1.0,
				},
			}

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestNewSlowInjector tests NewSlowInjector.
func TestNewSlowInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    time.Duration
		want    *SlowInjector
		wantErr error
	}{
		{
			give: 0,
			want: &SlowInjector{
				duration: 0,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			give: time.Millisecond,
			want: &SlowInjector{
				duration: time.Millisecond,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			give: time.Millisecond * 1000,
			want: &SlowInjector{
				duration: time.Second,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			give: time.Hour * 1000000,
			want: &SlowInjector{
				duration: time.Hour * 1000000,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
			t.Parallel()

			i, err := NewSlowInjector(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want.duration, i.duration)
		})
	}
}

// TestSlowInjectorHandler tests SlowInjector.Handler.
func TestSlowInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *SlowInjector
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
			give:     &SlowInjector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "valid",
			give: &SlowInjector{
				duration: time.Millisecond,
				sleep:    func(d time.Duration) { return },
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &Fault{
				opt: Options{
					Enabled:           true,
					Injector:          tt.give,
					PercentOfRequests: 1.0,
				},
			}

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}
