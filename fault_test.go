package fault

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

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
