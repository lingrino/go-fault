package fault

import (
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

			ci, err := NewChainInjector(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, len(tt.give), len(ci.middlewares))
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
				WithParticipation(1.0),
			)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}
