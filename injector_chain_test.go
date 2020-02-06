package fault

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
						})
					},
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							fmt.Fprint(w, "two")
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
