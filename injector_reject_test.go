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
				reporter: NewNoopReporter(),
			},
			wantErr: nil,
		},
		{
			name: "custom reporter",
			giveOptions: []RejectInjectorOption{
				WithReporter(newTestReporter()),
			},
			want: &RejectInjector{
				reporter: newTestReporter(),
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
		name        string
		giveOptions []RejectInjectorOption
	}{
		{
			name:        "valid",
			giveOptions: []RejectInjectorOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ri, err := NewRejectInjector(tt.giveOptions...)
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
