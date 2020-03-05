package fault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRejectInjector tests NewRejectInjector.
func TestNewRejectInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		giveOptions []RejectInjectorOption
		want        *RejectInjector
		wantErr     error
	}{
		{
			name:        "no options",
			giveOptions: []RejectInjectorOption{},
			want: &RejectInjector{
				reporter: &NoopReporter{},
			},
			wantErr: nil,
		},
		{
			name: "custom reporter",
			giveOptions: []RejectInjectorOption{
				WithReporter(newTestReporter()),
			},
			want: &RejectInjector{
				reporter: &testReporter{},
			},
			wantErr: nil,
		},
		{
			name: "option error",
			giveOptions: []RejectInjectorOption{
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

			ri, err := NewRejectInjector(tt.giveOptions...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, ri)
		})
	}
}

// TestRejectInjectorHandler tests RejectInjector.Handler.
func TestRejectInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "valid",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ri, err := NewRejectInjector()
			assert.NoError(t, err)

			f, err := NewFault(ri,
				WithEnabled(true),
				WithParticipation(1.0),
			)
			assert.NoError(t, err)

			rr := testRequestExpectPanic(t, f)
			assert.Nil(t, rr)
		})
	}
}
