package fault

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewChainInjector tests NewChainInjector.
func TestNewChainInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    []Injector
		wantErr error
	}{
		{
			name:    "nil",
			give:    nil,
			wantErr: nil,
		},
		{
			name:    "empty",
			give:    []Injector{},
			wantErr: nil,
		},
		{
			name: "one",
			give: []Injector{
				newTestInjectorNoop(),
			},
			wantErr: nil,
		},
		{
			name: "two",
			give: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewChainInjector(tt.give)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, len(tt.give), len(i.middlewares))
		})
	}
}

// TestChainInjectorHandler tests ChainInjector.Handler.
func TestChainInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     []Injector
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
			give:     []Injector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "one",
			give: []Injector{
				newTestInjectorOneOK(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "noop one",
			give: []Injector{
				newTestInjectorNoop(),
				newTestInjectorOneOK(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "two error",
			give: []Injector{
				newTestInjectorTwoTeapot(),
				newTestInjector500s(),
			},
			wantCode: http.StatusTeapot,
			wantBody: "two" + http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "one two",
			give: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorTwoTeapot(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + "two" + testHandlerBody,
		},
		{
			name: "one stop two",
			give: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorStop(),
				newTestInjectorTwoTeapot(),
			},
			wantCode: http.StatusOK,
			wantBody: "one",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ci, err := NewChainInjector(tt.give)
			assert.NoError(t, err)

			f, err := NewFault(ci,
				WithEnabled(true),
				WithInjectPercent(1.0),
			)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestNewRandomInjector tests NewRandomInjector.
func TestNewRandomInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		giveInjector []Injector
		giveOptions  []RandomInjectorOption
		wantRand     *rand.Rand
		wantErr      error
	}{
		{
			name:         "nil",
			giveInjector: nil,
			giveOptions:  nil,
			wantRand:     rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:      nil,
		},
		{
			name:         "empty",
			giveInjector: []Injector{},
			giveOptions:  nil,
			wantRand:     rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:      nil,
		},
		{
			name: "one",
			giveInjector: []Injector{
				newTestInjectorNoop(),
			},
			giveOptions: nil,
			wantRand:    rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:     nil,
		},
		{
			name: "two",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			giveOptions: nil,
			wantRand:    rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:     nil,
		},
		{
			name: "with seed",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			giveOptions: []RandomInjectorOption{
				WithRandSeed(100),
			},
			wantRand: rand.New(rand.NewSource(100)),
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewRandomInjector(tt.giveInjector, tt.giveOptions...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, len(tt.giveInjector), len(i.middlewares))
		})
	}
}

// TestRandomInjectorHandler tests RandomInjector.Handler.
func TestRandomInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     []Injector
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
			give:     []Injector{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "one",
			give: []Injector{
				newTestInjectorOneOK(),
			},
			wantCode: http.StatusOK,
			wantBody: "one" + testHandlerBody,
		},
		{
			name: "two",
			give: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorTwoTeapot(),
			},
			// With defaultRandSeed we will always choose the second item
			wantCode: http.StatusTeapot,
			wantBody: "two" + testHandlerBody,
		},
		{
			name: "seven",
			give: []Injector{
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorTwoTeapot(),
			},
			// With defaultRandSeed we will always choose the seventh item
			wantCode: http.StatusTeapot,
			wantBody: "two" + testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ri, err := NewRandomInjector(tt.give)
			assert.NoError(t, err)

			f, err := NewFault(ri,
				WithEnabled(true),
				WithInjectPercent(1.0),
			)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestNewRejectInjector tests NewRejectInjector.
func TestNewRejectInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    *RejectInjector
		wantErr error
	}{
		{
			name:    "valid",
			want:    &RejectInjector{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, err := NewRejectInjector()

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, i)
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
				WithInjectPercent(1.0),
			)
			assert.NoError(t, err)

			rr := testRequestExpectPanic(t, f)
			assert.Nil(t, rr)
		})
	}
}

// TestNewErrorInjector tests NewErrorInjector.
func TestNewErrorInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    []ErrorInjectorOption
		want    *ErrorInjector
		wantErr error
	}{
		{
			name:    "nil",
			give:    nil,
			want:    nil,
			wantErr: ErrInvalidHTTPCode,
		},
		{
			name: "only code",
			give: []ErrorInjectorOption{
				WithStatusCode(http.StatusCreated),
			},
			want: &ErrorInjector{
				statusCode: http.StatusCreated,
				statusText: http.StatusText(http.StatusCreated),
			},
			wantErr: nil,
		},
		{
			name: "code with different text",
			give: []ErrorInjectorOption{
				WithStatusCode(http.StatusCreated),
				WithStatusText(http.StatusText(http.StatusAccepted)),
			},
			want: &ErrorInjector{
				statusCode: http.StatusCreated,
				statusText: http.StatusText(http.StatusAccepted),
			},
			wantErr: nil,
		},
		{
			name: "code with random text",
			give: []ErrorInjectorOption{
				WithStatusCode(http.StatusTeapot),
				WithStatusText("wow very random"),
			},
			want: &ErrorInjector{
				statusCode: http.StatusTeapot,
				statusText: "wow very random",
			},
			wantErr: nil,
		},
		{
			name: "invalid code",
			give: []ErrorInjectorOption{
				WithStatusCode(0),
				WithStatusText("invalid code"),
			},
			want:    nil,
			wantErr: ErrInvalidHTTPCode,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.name), func(t *testing.T) {
			t.Parallel()

			i, err := NewErrorInjector(tt.give...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, i)
		})
	}
}

// TestErrorInjectorHandler tests ErrorInjector.Handler.
func TestErrorInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     []ErrorInjectorOption
		wantCode int
		wantBody string
	}{
		{
			name: "only code",
			give: []ErrorInjectorOption{
				WithStatusCode(http.StatusInternalServerError),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "custom text",
			give: []ErrorInjectorOption{
				WithStatusCode(http.StatusInternalServerError),
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

			ei, err := NewErrorInjector(tt.give...)
			assert.NoError(t, err)

			f, err := NewFault(ei,
				WithEnabled(true),
				WithInjectPercent(1.0),
			)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}

// TestNewSlowInjector tests NewSlowInjector.
func TestNewSlowInjector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		give    []SlowInjectorOption
		want    *SlowInjector
		wantErr error
	}{
		{
			name: "nil",
			give: nil,
			want: &SlowInjector{
				duration: 0,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			name: "empty",
			give: []SlowInjectorOption{},
			want: &SlowInjector{
				duration: 0,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			name: "custom duration",
			give: []SlowInjectorOption{
				WithDuration(time.Minute),
			},
			want: &SlowInjector{
				duration: time.Minute,
				sleep:    time.Sleep,
			},
			wantErr: nil,
		},
		{
			name: "custom sleep",
			give: []SlowInjectorOption{
				WithDuration(time.Minute),
				WithSleepFunction(func(time.Duration) {}),
			},
			want: &SlowInjector{
				duration: time.Minute,
				sleep:    func(time.Duration) {},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%v", tt.give), func(t *testing.T) {
			t.Parallel()

			si, err := NewSlowInjector(tt.give...)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want.duration, si.duration)
		})
	}
}

// TestSlowInjectorHandler tests SlowInjector.Handler.
func TestSlowInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		give     []SlowInjectorOption
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
			give:     []SlowInjectorOption{},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "with time.Sleep",
			give: []SlowInjectorOption{
				WithDuration(time.Nanosecond),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
		{
			name: "with custom function",
			give: []SlowInjectorOption{
				WithDuration(time.Hour),
				WithSleepFunction(func(time.Duration) {}),
			},
			wantCode: testHandlerCode,
			wantBody: testHandlerBody,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			si, err := NewSlowInjector(tt.give...)
			assert.NoError(t, err)

			f, err := NewFault(si,
				WithEnabled(true),
				WithInjectPercent(1.0),
			)
			assert.NoError(t, err)

			rr := testRequest(t, f)

			assert.Equal(t, tt.wantCode, rr.Code)
			assert.Equal(t, tt.wantBody, strings.TrimSpace(rr.Body.String()))
		})
	}
}
