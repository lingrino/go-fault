package fault

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewErrorInjector tests NewErrorInjector.
func TestNewErrorInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		giveCode    int
		giveOptions []ErrorInjectorOption
		want        *ErrorInjector
		wantErr     error
	}{
		{
			name:        "only code",
			giveCode:    http.StatusCreated,
			giveOptions: nil,
			want: &ErrorInjector{
				statusCode: http.StatusCreated,
				statusText: http.StatusText(http.StatusCreated),
				reporter:   NewNoopReporter(),
			},
			wantErr: nil,
		},
		{
			name:     "code with different text",
			giveCode: http.StatusCreated,
			giveOptions: []ErrorInjectorOption{
				WithStatusText(http.StatusText(http.StatusAccepted)),
			},
			want: &ErrorInjector{
				statusCode: http.StatusCreated,
				statusText: http.StatusText(http.StatusAccepted),
				reporter:   NewNoopReporter(),
			},
			wantErr: nil,
		},
		{
			name:     "code with random text",
			giveCode: http.StatusTeapot,
			giveOptions: []ErrorInjectorOption{
				WithStatusText("wow very random"),
			},
			want: &ErrorInjector{
				statusCode: http.StatusTeapot,
				statusText: "wow very random",
				reporter:   NewNoopReporter(),
			},
			wantErr: nil,
		},
		{
			name:     "custom reporter",
			giveCode: http.StatusOK,
			giveOptions: []ErrorInjectorOption{
				WithReporter(newTestReporter()),
			},
			want: &ErrorInjector{
				statusCode: http.StatusOK,
				statusText: http.StatusText(http.StatusOK),
				reporter:   newTestReporter(),
			},
			wantErr: nil,
		},
		{
			name:     "invalid code",
			giveCode: 0,
			giveOptions: []ErrorInjectorOption{
				WithStatusText("invalid code"),
			},
			want:    nil,
			wantErr: ErrInvalidHTTPCode,
		},
		{
			name:     "option error",
			giveCode: 200,
			giveOptions: []ErrorInjectorOption{
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

			ei, err := NewErrorInjector(tt.giveCode, tt.giveOptions...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, ei)
		})
	}
}

// TestErrorInjectorHandler tests ErrorInjector.Handler.
func TestErrorInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		giveCode    int
		giveOptions []ErrorInjectorOption
		wantCode    int
		wantBody    string
	}{
		{
			name:        "only code",
			giveCode:    http.StatusInternalServerError,
			giveOptions: nil,
			wantCode:    http.StatusInternalServerError,
			wantBody:    http.StatusText(http.StatusInternalServerError),
		},
		{
			name:     "custom text",
			giveCode: http.StatusInternalServerError,
			giveOptions: []ErrorInjectorOption{
				WithStatusText("very custom text"),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "very custom text",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ei, err := NewErrorInjector(tt.giveCode, tt.giveOptions...)
			assert.NoError(t, err)

			f, err := NewFault(ei,
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
