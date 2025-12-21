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
		name         string
		giveInjector []Injector
		wantErr      error
	}{
		{
			name:         "nil",
			giveInjector: nil,
			wantErr:      nil,
		},
		{
			name:         "nil injector in slice",
			giveInjector: []Injector{nil},
			wantErr:      ErrNilInjector,
		},
		{
			name:         "empty",
			giveInjector: []Injector{},
			wantErr:      nil,
		},
		{
			name: "one",
			giveInjector: []Injector{
				newTestInjectorNoop(t),
			},
			wantErr: nil,
		},
		{
			name: "two",
			giveInjector: []Injector{
				newTestInjectorNoop(t),
				newTestInjector500s(t),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ci, err := NewChainInjector(tt.giveInjector)

			assert.Equal(t, tt.wantErr, err)

			if tt.wantErr == nil {
				assert.Equal(t, len(tt.giveInjector), len(ci.middlewares))
			} else {
				assert.Nil(t, ci)
			}
		})
	}
}

// TestChainInjectorHandler tests ChainInjector.Handler.
func TestChainInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveInjector []Injector
		wantCode     int
		wantBody     string
	}{
		{
			name:         "nil",
			giveInjector: nil,
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name:         "empty",
			giveInjector: []Injector{},
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name: "one",
			giveInjector: []Injector{
				newTestInjectorOneOK(t),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "noop one",
			giveInjector: []Injector{
				newTestInjectorNoop(t),
				newTestInjectorOneOK(t),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "two error",
			giveInjector: []Injector{
				newTestInjectorTwoTeapot(t),
				newTestInjector500s(t),
			},
			wantCode: http.StatusTeapot,
			wantBody: "two" + http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "one two",
			giveInjector: []Injector{
				newTestInjectorOneOK(t),
				newTestInjectorTwoTeapot(t),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + "two" + testHandlerBody,
		},
		{
			name: "one stop two",
			giveInjector: []Injector{
				newTestInjectorOneOK(t),
				newTestInjectorStop(t),
				newTestInjectorTwoTeapot(t),
			},
			wantCode: http.StatusOK,
			wantBody: "one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ci, err := NewChainInjector(tt.giveInjector)
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
