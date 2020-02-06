package fault

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
