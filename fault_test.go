package fault

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewFault tests NewFault.
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
			giveInjector: newTestInjectorNoop(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathBlocklist([]string{"/donotinject"}),
				WithPathAllowlist([]string{"/onlyinject"}),
				WithHeaderBlocklist(map[string]string{"block": "yes"}),
				WithHeaderAllowlist(map[string]string{"allow": "yes"}),
				WithRandSeed(100),
				WithRandFloat32Func(func() float32 { return 0.0 }),
			},
			wantFault: &Fault{
				enabled:       true,
				injector:      newTestInjectorNoop(t),
				participation: 1.0,
				pathBlocklist: map[string]bool{
					"/donotinject": true,
				},
				pathAllowlist: map[string]bool{
					"/onlyinject": true,
				},
				headerBlocklist: map[string]string{
					"block": "yes",
				},
				headerAllowlist: map[string]string{
					"allow": "yes",
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
			giveInjector: newTestInjectorNoop(t),
			giveOptions: []Option{
				WithParticipation(100.0),
			},
			wantFault: nil,
			wantErr:   ErrInvalidPercent,
		},
		{
			name:         "option error",
			giveInjector: newTestInjectorNoop(t),
			giveOptions: []Option{
				withError(),
			},
			wantFault: nil,
			wantErr:   errErrorOption,
		},
		{
			name:         "empty options",
			giveInjector: newTestInjectorNoop(t),
			giveOptions:  []Option{},
			wantFault: &Fault{
				enabled:       false,
				injector:      newTestInjectorNoop(t),
				participation: 0.0,
				pathBlocklist: nil,
				pathAllowlist: nil,
				randSeed:      defaultRandSeed,
				rand:          rand.New(rand.NewSource(defaultRandSeed)),
				randF:         rand.New(rand.NewSource(defaultRandSeed)).Float32,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(tt.giveInjector, tt.giveOptions...)

			// Function equality cannot be determined so set to nil before comparing
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
			giveInjector: newTestInjectorNoop(t),
			giveOptions: []Option{
				WithEnabled(false),
				WithParticipation(1.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "zero percent",
			giveInjector: newTestInjectorNoop(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(0.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "0 percent 500s custom function",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithRandFloat32Func(func() float32 { return 1.0 }),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with blocklist root",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathBlocklist([]string{"/"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with allowlist root",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathAllowlist([]string{"/"}),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "100 percent 500s with allowlist other",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathAllowlist([]string{"/onlyinject"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with allowlist and blocklist root",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithPathBlocklist([]string{"/"}),
				WithPathAllowlist([]string{"/"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with header block",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithHeaderBlocklist(map[string]string{testHeaderKey: testHeaderVal}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with header allow",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithHeaderAllowlist(map[string]string{testHeaderKey: testHeaderVal}),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name:         "100 percent 500s with header allow other",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithHeaderAllowlist(map[string]string{"header": "not in test request"}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent 500s with header allowlist and blocklist",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
				WithHeaderBlocklist(map[string]string{testHeaderKey: testHeaderVal}),
				WithHeaderAllowlist(map[string]string{testHeaderKey: testHeaderVal}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "disabled with with path/header allowlists",
			giveInjector: newTestInjector500s(t),
			giveOptions: []Option{
				WithEnabled(false),
				WithParticipation(1.0),
				WithPathAllowlist([]string{"/"}),
				WithHeaderAllowlist(map[string]string{testHeaderKey: testHeaderVal}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name:         "100 percent inject nothing",
			giveInjector: newTestInjectorNoop(t),
			giveOptions: []Option{
				WithEnabled(true),
				WithParticipation(1.0),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
	}

	for _, tt := range tests {
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

// TestFaultSetEnabled tests Fault.SetEnabled().
func TestFaultSetEnabled(t *testing.T) {
	t.Parallel()

	f, err := NewFault(newTestInjector500s(t),
		WithEnabled(true),
		WithParticipation(1.0),
	)
	assert.NoError(t, err)

	rr := testRequest(t, f)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), strings.TrimSpace(rr.Body.String()))

	f.SetEnabled(false)

	rr = testRequest(t, f)
	assert.Equal(t, testHandlerCode, rr.Code)
	assert.Equal(t, testHandlerBody, strings.TrimSpace(rr.Body.String()))
}

// TestFaultSetParticipation tests Fault.SetParticipation().
func TestFaultSetParticipation(t *testing.T) {
	t.Parallel()

	f, err := NewFault(newTestInjector500s(t),
		WithEnabled(true),
		WithParticipation(1.0),
	)
	assert.NoError(t, err)

	rr := testRequest(t, f)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), strings.TrimSpace(rr.Body.String()))

	err = f.SetParticipation(0.0)
	assert.NoError(t, err)

	rr = testRequest(t, f)
	assert.Equal(t, testHandlerCode, rr.Code)
	assert.Equal(t, testHandlerBody, strings.TrimSpace(rr.Body.String()))

	// Test invalid participation values
	err = f.SetParticipation(-0.1)
	assert.Equal(t, ErrInvalidPercent, err)

	err = f.SetParticipation(1.1)
	assert.Equal(t, ErrInvalidPercent, err)
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
		t.Run(fmt.Sprintf("%g", tt.givePercent), func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(newTestInjectorNoop(t),
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

// TestFaultConcurrentAccess verifies that SetEnabled and SetParticipation
// are safe to call concurrently with Handler.
func TestFaultConcurrentAccess(t *testing.T) {
	t.Parallel()

	f, err := NewFault(newTestInjectorNoop(t),
		WithEnabled(true),
		WithParticipation(0.5),
	)
	assert.NoError(t, err)

	handler := f.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	const iterations = 1000

	// Concurrently toggle enabled
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			f.SetEnabled(i%2 == 0)
		}
	}()

	// Concurrently change participation
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			err := f.SetParticipation(float32(i%100) / 100.0)
			assert.NoError(t, err)
		}
	}()

	// Concurrently call Handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			req := httptest.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
	}()

	wg.Wait()
}
