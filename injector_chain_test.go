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
		giveOptions  []ChainInjectorOption
		wantErr      error
	}{
		{
			name:         "nil",
			giveInjector: nil,
			giveOptions:  []ChainInjectorOption{},
			wantErr:      nil,
		},
		{
			name:         "empty",
			giveInjector: []Injector{},
			giveOptions:  []ChainInjectorOption{},
			wantErr:      nil,
		},
		{
			name: "one",
			giveInjector: []Injector{
				newTestInjectorNoop(),
			},
			giveOptions: []ChainInjectorOption{},
			wantErr:     nil,
		},
		{
			name: "two",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			giveOptions: []ChainInjectorOption{},
			wantErr:     nil,
		},
		{
			name: "option error",
			giveInjector: []Injector{
				newTestInjectorNoop(),
			},
			giveOptions: []ChainInjectorOption{
				withError(),
			},
			wantErr: errErrorOption,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ci, err := NewChainInjector(tt.giveInjector, tt.giveOptions...)

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
		giveOptions  []ChainInjectorOption
		wantCode     int
		wantBody     string
	}{
		{
			name:         "nil",
			giveInjector: nil,
			giveOptions:  []ChainInjectorOption{},
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name:         "empty",
			giveInjector: []Injector{},
			giveOptions:  []ChainInjectorOption{},
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name: "one",
			giveInjector: []Injector{
				newTestInjectorOneOK(),
			},
			giveOptions: []ChainInjectorOption{},
			wantCode:    http.StatusOK,
			wantBody:    "one" + testHandlerBody,
		},
		{
			name: "noop one",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjectorOneOK(),
			},
			giveOptions: []ChainInjectorOption{},
			wantCode:    http.StatusOK,
			wantBody:    "one" + testHandlerBody,
		},
		{
			name: "two error",
			giveInjector: []Injector{
				newTestInjectorTwoTeapot(),
				newTestInjector500s(),
			},
			giveOptions: []ChainInjectorOption{},
			wantCode:    http.StatusTeapot,
			wantBody:    "two" + http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "one two",
			giveInjector: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorTwoTeapot(),
			},
			giveOptions: []ChainInjectorOption{},
			wantCode:    http.StatusOK,
			wantBody:    "one" + "two" + testHandlerBody,
		},
		{
			name: "one stop two",
			giveInjector: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorStop(),
				newTestInjectorTwoTeapot(),
			},
			giveOptions: []ChainInjectorOption{},
			wantCode:    http.StatusOK,
			wantBody:    "one",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ci, err := NewChainInjector(tt.giveInjector, tt.giveOptions...)
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
