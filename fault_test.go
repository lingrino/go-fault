package fault

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		give      Options
		wantFault *Fault
		wantErr   bool
	}{
		{
			name: "valid",
			give: Options{
				Enabled:           true,
				Injector:          newTestInjector(false),
				PercentOfRequests: 1.0,
			},
			wantFault: &Fault{
				opt: Options{
					Enabled: true,
					Injector: &testInjector{
						resp500: false,
					},
					PercentOfRequests: 1.0,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid injector",
			give: Options{
				Injector:          nil,
				PercentOfRequests: 1.0,
			},
			wantFault: nil,
			wantErr:   true,
		},
		{
			name: "invalid percent",
			give: Options{
				Injector:          newTestInjector(false),
				PercentOfRequests: 1.1,
			},
			wantFault: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := NewFault(tt.give)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantFault, f)
		})
	}
}

func TestFaultHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *Fault
		wantCode int
		wantBody string
	}{
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

func TestFaultPercentDo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		givePercent float32
		wantRange   float32
	}{
		{1.0, 0},
		{0.75, 0.005},
		{0.3298, 0.005},
		{0.0001, 0.005},
		{0.0, 0.0},
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

			minP := tt.givePercent - tt.wantRange
			per := errorC / totalC
			maxP := tt.givePercent + tt.wantRange

			if per < minP || per > maxP {
				t.Errorf("wrong distribution. expected: %v < per < %v, got: %v", minP, maxP, per)
			}
		})
	}
}

func TestNewRejectInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    *RejectInjector
		wantErr bool
	}{
		{
			name:    "valid",
			want:    &RejectInjector{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewRejectInjector()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, i)
		})
	}
}

func TestRejectInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		give *RejectInjector
	}{
		{
			name: "valid",
			give: &RejectInjector{},
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

			rr := testRequestExpectPanic(t, f)
			assert.Nil(t, rr)
		})
	}
}

func TestNewErrorInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    int
		want    *ErrorInjector
		wantErr bool
	}{
		{
			give:    -1,
			want:    nil,
			wantErr: true,
		},
		{
			give:    0,
			want:    nil,
			wantErr: true,
		},
		{
			give: 200,
			want: &ErrorInjector{
				statusCode: 200,
				statusText: http.StatusText(200),
			},
			wantErr: false,
		},
		{
			give: 500,
			want: &ErrorInjector{
				statusCode: 500,
				statusText: http.StatusText(500),
			},
			wantErr: false,
		},
		{
			give:    120000,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
			t.Parallel()

			i, err := NewErrorInjector(tt.give)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, i)

		})
	}
}
func TestErrorInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *ErrorInjector
		wantCode int
		wantBody string
	}{
		{
			name: "200",
			give: &ErrorInjector{
				statusCode: 200,
				statusText: http.StatusText(200),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "418",
			give: &ErrorInjector{
				statusCode: 418,
				statusText: http.StatusText(418),
			},
			wantCode: 418,
			wantBody: http.StatusText(418),
		},
		{
			name: "500",
			give: &ErrorInjector{
				statusCode: 500,
				statusText: http.StatusText(500),
			},
			wantCode: 500,
			wantBody: http.StatusText(500),
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

func TestNewSlowInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    time.Duration
		want    *SlowInjector
		wantErr bool
	}{
		{
			give: time.Millisecond,
			want: &SlowInjector{
				duration: time.Millisecond,
				sleep:    time.Sleep,
			},
			wantErr: false,
		},
		{
			give: time.Millisecond * 1000,
			want: &SlowInjector{
				duration: time.Second,
				sleep:    time.Sleep,
			},
			wantErr: false,
		},
		{
			give: time.Hour * 1000000,
			want: &SlowInjector{
				duration: time.Hour * 1000000,
				sleep:    time.Sleep,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
			t.Parallel()

			i, err := NewSlowInjector(tt.give)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.duration, i.duration)
		})
	}
}

func TestSlowInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *SlowInjector
		wantCode int
		wantBody string
	}{
		{
			name: "valid",
			give: &SlowInjector{
				duration: time.Millisecond,
				sleep:    func(d time.Duration) { return },
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
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
