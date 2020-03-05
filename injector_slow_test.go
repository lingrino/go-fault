package fault

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewSlowInjector tests NewSlowInjector.
func TestNewSlowInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveDuration time.Duration
		giveOptions  []SlowInjectorOption
		want         *SlowInjector
		wantErr      error
	}{
		{
			name:         "nil",
			giveDuration: 0,
			giveOptions:  nil,
			want: &SlowInjector{
				duration: 0,
				slowF:    time.Sleep,
				reporter: &NoopReporter{},
			},
			wantErr: nil,
		},
		{
			name:         "empty",
			giveDuration: 0,
			giveOptions:  []SlowInjectorOption{},
			want: &SlowInjector{
				duration: 0,
				slowF:    time.Sleep,
				reporter: &NoopReporter{},
			},
			wantErr: nil,
		},
		{
			name:         "custom duration",
			giveDuration: time.Minute,
			giveOptions:  nil,
			want: &SlowInjector{
				duration: time.Minute,
				slowF:    time.Sleep,
				reporter: &NoopReporter{},
			},
			wantErr: nil,
		},
		{
			name:         "custom sleep",
			giveDuration: time.Minute,
			giveOptions: []SlowInjectorOption{
				WithSlowFunc(func(time.Duration) {}),
			},
			want: &SlowInjector{
				duration: time.Minute,
				slowF:    func(time.Duration) {},
				reporter: &NoopReporter{},
			},
			wantErr: nil,
		},
		{
			name:         "custom reporter",
			giveDuration: time.Minute,
			giveOptions: []SlowInjectorOption{
				WithReporter(newTestReporter()),
			},
			want: &SlowInjector{
				duration: time.Minute,
				slowF:    time.Sleep,
				reporter: &testReporter{},
			},
			wantErr: nil,
		},
		{
			name:         "option error",
			giveDuration: time.Minute,
			giveOptions: []SlowInjectorOption{
				withError(),
			},
			want:    nil,
			wantErr: errErrorOption,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			si, err := NewSlowInjector(tt.giveDuration, tt.giveOptions...)

			// Function equality cannot be determined so we set these to nil before
			// doing our comparison
			if tt.want != nil {
				si.slowF = nil
				tt.want.slowF = nil
			}

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, si)
		})
	}
}

// TestSlowInjectorHandler tests SlowInjector.Handler.
func TestSlowInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveDuration time.Duration
		giveOptions  []SlowInjectorOption
		wantCode     int
		wantBody     string
	}{
		{
			name:         "nil",
			giveDuration: 0,
			giveOptions:  nil,
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name:         "empty",
			giveDuration: 0,
			giveOptions:  []SlowInjectorOption{},
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name:         "with time.Sleep",
			giveDuration: time.Microsecond,
			giveOptions:  nil,
			wantCode:     testHandlerCode,
			wantBody:     testHandlerBody,
		},
		{
			name:         "with custom function",
			giveDuration: time.Hour,
			giveOptions: []SlowInjectorOption{
				WithSlowFunc(func(time.Duration) {}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			si, err := NewSlowInjector(tt.giveDuration, tt.giveOptions...)
			assert.NoError(t, err)

			f, err := NewFault(si,
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
