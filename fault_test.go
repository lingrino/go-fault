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
		giveOptions  []Option
		wantFault    *Fault
		wantErr      error
	}{
		{
			name:         "all options",
			giveInjector: newTestInjectorNoop(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathBlacklist([]string{"/donotinject"}),
				WithPathWhitelist([]string{"/onlyinject"}),
				WithRandSeed(100),
				WithRandFloat32Func(func() float32 { return 0.0 }),
			},
			wantFault: &Fault{
				enabled:       true,
				injector:      &testInjectorNoop{},
				participation: 1.0,
				pathBlacklist: map[string]bool{
					"/donotinject": true,
				},
				pathWhitelist: map[string]bool{
					"/onlyinject": true,
				},
				randSeed: 100,
				rand:     rand.New(rand.NewSource(100)),
				randF:    func() float32 { return 0.0 },
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
			giveInjector: newTestInjectorNoop(),
			giveOptions: []Option{
				WithParticipation(100.0),
			},
			wantFault: nil,
			wantErr:   ErrInvalidPercent,
		},
		{
			name:         "option error",
			giveInjector: newTestInjectorNoop(),
			giveOptions: []Option{
				withError(),
			},
			wantFault: nil,
			wantErr:   errErrorOption,
		},
		{
			name:         "empty options",
			giveInjector: newTestInjectorNoop(),
			giveOptions:  []Option{},
			wantFault: &Fault{
				enabled:       false,
				injector:      &testInjectorNoop{},
				participation: 0.0,
				pathBlacklist: nil,
				pathWhitelist: nil,
				randSeed:      defaultRandSeed,
				rand:          rand.New(rand.NewSource(defaultRandSeed)),
				randF:         rand.New(rand.NewSource(defaultRandSeed)).Float32,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(tt.giveInjector, tt.giveOptions...)

			// Function equality cannot be determined so we set these to nil before doing our comparison
			if tt.wantFault != nil {
				f.randF = nil
				tt.wantFault.randF = nil
			}

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
		giveOptions  []Option
		wantCode     int
		wantBody     string
	}{
		{
			name:         "not enabled",
			giveInjector: newTestInjectorNoop(),
			giveOptions: []Option{
				WithEnabled(false),
				WithParticipation(1.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "zero percent",
			giveInjector: newTestInjectorNoop(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(0.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s",
			giveInjector: newTestInjector500s(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "0 percent 500s custom function",
			giveInjector: newTestInjector500s(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithRandFloat32Func(func() float32 { return 1.0 }),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with blacklist root",
			giveInjector: newTestInjector500s(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathBlacklist([]string{"/"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with whitelist root",
			giveInjector: newTestInjector500s(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathWhitelist([]string{"/"}),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "100 percent 500s with whitelist other",
			giveInjector: newTestInjector500s(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathWhitelist([]string{"/onlyinject"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with whitelist and blacklist root",
			giveInjector: newTestInjector500s(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathBlacklist([]string{"/"}),
				WithPathWhitelist([]string{"/"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent inject nothing",
			giveInjector: newTestInjectorNoop(),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
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

// TestFaultPercentDo tests the internal Fault.participate().
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

			f, err := NewFault(newTestInjectorNoop(),
				WithParticipation(tt.givePercent),
			)
			assert.NoError(t, err)

			var trueC, totalC float32
			for totalC <= 100000 {
				result := f.participate()
				if result {
					trueC++
				}
				totalC++
			}

			minP := tt.wantPercent - tt.wantRange
			per := trueC / totalC
			maxP := tt.wantPercent + tt.wantRange

			assert.GreaterOrEqual(t, per, minP)
			assert.LessOrEqual(t, per, maxP)
		})
	}
}
