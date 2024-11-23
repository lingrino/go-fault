package fault

import (
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{
			name: "with custom function",
			giveInjector: []Injector{
				newTestInjectorNoop(),
				newTestInjector500s(),
			},
			giveOptions: []RandomInjectorOption{
				WithRandIntFunc(func(int) int { return 1 }),
			},
			wantRand: rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:  nil,
		},
		{
			name: "option error",
			giveInjector: []Injector{
				newTestInjectorNoop(),
			},
			giveOptions: []RandomInjectorOption{
				withError(),
			},
			wantRand: rand.New(rand.NewSource(defaultRandSeed)),
			wantErr:  errErrorOption,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ri, err := NewRandomInjector(tt.giveInjector, tt.giveOptions...)

			assert.Equal(t, tt.wantErr, err)

			if tt.wantErr == nil {
				assert.Equal(t, tt.wantRand, ri.rand)
				assert.Equal(t, len(tt.giveInjector), len(ri.middlewares))
			} else {
				assert.Nil(t, ri)
			}
		})
	}
}

// TestRandomInjectorHandler tests RandomInjector.Handler.
func TestRandomInjectorHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		give        []Injector
		giveOptions []RandomInjectorOption
		wantCode    int
		wantBody    string
	}{
		{
			name:        "nil",
			give:        nil,
			giveOptions: nil,
			wantCode:    testHandlerCode,
			wantBody:    testHandlerBody,
		},
		{
			name:        "empty",
			give:        []Injector{},
			giveOptions: nil,
			wantCode:    testHandlerCode,
			wantBody:    testHandlerBody,
		},
		{
			name: "one",
			give: []Injector{
				newTestInjectorOneOK(),
			},
			giveOptions: nil,
			wantCode:    http.StatusOK,
			wantBody:    "one" + testHandlerBody,
		},
		{
			name: "two",
			give: []Injector{
				newTestInjectorOneOK(),
				newTestInjectorTwoTeapot(),
			},
			giveOptions: nil,
			// defaultRandSeed will choose 1
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
			giveOptions: nil,
			// defaultRandSeed will choose 6
			wantCode: http.StatusTeapot,
			wantBody: "two" + testHandlerBody,
		},
		{
			name: "custom rand func",
			give: []Injector{
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorTwoTeapot(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
				newTestInjectorNoop(),
			},
			giveOptions: []RandomInjectorOption{
				WithRandIntFunc(func(int) int { return 2 }),
			},
			// defaultRandSeed will choose 6. Custom function should choose 2.
			wantCode: http.StatusTeapot,
			wantBody: "two" + testHandlerBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ri, err := NewRandomInjector(tt.give, tt.giveOptions...)
			assert.NoError(t, err)

			f, err := NewFault(ri,
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
