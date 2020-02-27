package fault

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewFault tests NewFault().
func TestNewFault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveInjector Injector
		giveOptions  []FaultOption
		wantFault    *Fault
		wantErr      error
	}{
		{
			name:         "all options",
			giveInjector: newTestInjector(false),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
				WithPathBlacklist([]string{"/donotinject"}),
				WithPathWhitelist([]string{"/onlyinject"}),
				WithRandSeed(100),
			},
			wantFault: &Fault{
				enabled: true,
				injector: &testInjector{
					resp500: false,
				},
				injectPercent: 1.0,
				pathBlacklist: map[string]bool{
					"/donotinject": true,
				},
				pathWhitelist: map[string]bool{
					"/onlyinject": true,
				},
				randSeed: 100,
				rand:     rand.New(rand.NewSource(100)),
			},
			wantErr: nil,
		},
		{
			name:         "nil injector",
			giveInjector: nil,
			giveOptions:  nil,
			wantFault:    nil,
			wantErr:      ErrNilInjector,
		},
		{
			name:         "invalid percent",
			giveInjector: newTestInjector(false),
			giveOptions: []FaultOption{
				WithInjectPercent(100.0),
			},
			wantFault: nil,
			wantErr:   ErrInvalidPercent,
		},
		{
			name:         "empty options",
			giveInjector: newTestInjector(false),
			giveOptions:  []FaultOption{},
			wantFault: &Fault{
				enabled: false,
				injector: &testInjector{
					resp500: false,
				},
				injectPercent: 0.0,
				pathBlacklist: nil,
				pathWhitelist: nil,
				randSeed:      defaultRandSeed,
				rand:          rand.New(rand.NewSource(defaultRandSeed)),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(tt.giveInjector, tt.giveOptions...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantFault, f)
		})
	}
}

// TestFaultHandler tests Fault.Handler.
func TestFaultHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveInjector Injector
		giveOptions  []FaultOption
		wantCode     int
		wantBody     string
	}{
		{
			name:         "not enabled",
			giveInjector: newTestInjector(false),
			giveOptions: []FaultOption{
				WithEnabled(false),
				WithInjectPercent(1.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "zero percent",
			giveInjector: newTestInjector(false),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(0.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s",
			giveInjector: newTestInjector(true),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "100 percent 500s with blacklist root",
			giveInjector: newTestInjector(true),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
				WithPathBlacklist([]string{"/"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with whitelist root",
			giveInjector: newTestInjector(true),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
				WithPathWhitelist([]string{"/"}),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "100 percent 500s with whitelist other",
			giveInjector: newTestInjector(true),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
				WithPathWhitelist([]string{"/onlyinject"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with whitelist and blacklist root",
			giveInjector: newTestInjector(true),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
				WithPathBlacklist([]string{"/"}),
				WithPathWhitelist([]string{"/"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent inject nothing",
			giveInjector: newTestInjector(false),
			giveOptions: []FaultOption{
				WithEnabled(true),
				WithInjectPercent(1.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(tt.giveInjector, tt.giveOptions...)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestFaultPercentDo tests the internal Fault.percentDo().
func TestFaultPercentDo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		givePercent float32
		wantPercent float32
		wantRange   float32
	}{
		{},
		{0.0, 0.0, 0.0},
		{0.0001, 0.0001, 0.005},
		{0.3298, 0.3298, 0.005},
		{0.75, 0.75, 0.005},
		{1.0, 1.0, 0.0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%g", tt.givePercent), func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(newTestInjector(false), WithInjectPercent(tt.givePercent))
			assert.NoError(t, err)

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
