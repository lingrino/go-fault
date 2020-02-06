package fault

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewSlowInjector tests NewSlowInjector.
func TestNewSlowInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    time.Duration
		want    *SlowInjector
		wantErr error
	}{
		{
			give: 0,
			want: &SlowInjector{
				duration: 0,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			give: time.Millisecond,
			want: &SlowInjector{
				duration: time.Millisecond,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			give: time.Millisecond * 1000,
			want: &SlowInjector{
				duration: time.Second,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			give: time.Hour * 1000000,
			want: &SlowInjector{
				duration: time.Hour * 1000000,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
			t.Parallel()

			i, err := NewSlowInjector(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want.duration, i.duration)
		})
	}
}

// TestSlowInjectorHandler tests SlowInjector.Handler.
func TestSlowInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     *SlowInjector
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
			give:     &SlowInjector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "valid",
			give: &SlowInjector{
				duration: time.Millisecond,
				sleep:    func(d time.Duration) {},
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
