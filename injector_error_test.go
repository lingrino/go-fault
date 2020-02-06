package fault

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
