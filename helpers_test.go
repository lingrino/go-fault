package fault

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// testHandlerCode and testHandlerBody are the default status code and status text expected
	// from a handler that has not been changed by an Injector. Don't use http.StatusOK because
	// some http methods default to http.StatusOK and then there's no difference between our
	// test response and other standard responses.
	testHandlerCode = http.StatusAccepted
	testHandlerBody = "Accepted"
	testHeaderKey   = "testing header key"
	testHeaderVal   = "testing header val"
)

// testRequest simulates a request to testHandler with a Fault injected.
func testRequest(t *testing.T, f *Fault) *httptest.ResponseRecorder {
	t.Helper()

	var testHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, testHandlerBody, testHandlerCode)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(testHeaderKey, testHeaderVal)

	rr := httptest.NewRecorder()

	if f != nil {
		finalHandler := f.Handler(testHandler)
		finalHandler.ServeHTTP(rr, req)
	} else {
		testHandler.ServeHTTP(rr, req)
	}

	return rr
}

// testRequestExpectPanic runs testRequest and catches/passes if panic(http.ErrAbortHandler).
func testRequestExpectPanic(t *testing.T, f *Fault) *httptest.ResponseRecorder {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			if !errors.Is(r.(error), http.ErrAbortHandler) {
				t.Fatal(r)
			}
		}
	}()

	rr := testRequest(t, f)

	return rr
}

// testInjectorNoop is an injector that does nothing.
type testInjectorNoop struct {
	t *testing.T
}

// newTestInjectorNoop creates a new testInjectorNoop.
func newTestInjectorNoop(t *testing.T) *testInjectorNoop {
	return &testInjectorNoop{t: t}
}

// Handler does nothing.
func (i *testInjectorNoop) Handler(next http.Handler) http.Handler { return next }

// testInjectorStop is an injector that stops a request.
type testInjectorStop struct {
	t *testing.T
}

// newTestInjectorStop creates a new testInjectorStop.
func newTestInjectorStop(t *testing.T) *testInjectorStop {
	return &testInjectorStop{t: t}
}

// Handler returns a Handler that stops the request.
func (i *testInjectorStop) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

// testInjector500s is an injector that returns 500s.
type testInjector500s struct {
	t *testing.T
}

// newTestInjector500 creates a new testInjector500s.
func newTestInjector500s(t *testing.T) *testInjector500s {
	return &testInjector500s{t: t}
}

// Handler returns a 500.
func (i *testInjector500s) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})
}

// testInjectorOneOK is an injector that writes "one" and statusOK.
type testInjectorOneOK struct {
	t *testing.T
}

// newTestInjectorOneOK creates a new testInjectorOneOK.
func newTestInjectorOneOK(t *testing.T) *testInjectorOneOK {
	return &testInjectorOneOK{t: t}
}

// Handler writes statusOK and "one" and continues.
func (i *testInjectorOneOK) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, err := fmt.Fprint(w, "one")
		assert.NoError(i.t, err)

		next.ServeHTTP(w, r)
	})
}

// testInjectorTwoTeapot is an injector that writes "two" and statusTeapot.
type testInjectorTwoTeapot struct {
	t *testing.T
}

// newTestInjectorTwoTeapot creates a new testInjectorTwoTeapot.
func newTestInjectorTwoTeapot(t *testing.T) *testInjectorTwoTeapot {
	return &testInjectorTwoTeapot{t: t}
}

// Handler writes StatusTeapot and "two" and continues.
func (i *testInjectorTwoTeapot) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)

		_, err := fmt.Fprint(w, "two")
		assert.NoError(i.t, err)

		next.ServeHTTP(w, r)
	})
}

var (
	errErrorOption = errors.New("intentional error for tests")
)

// errorOption returns errErrorOption.
type errorOption interface {
	Option
	ChainInjectorOption
	RandomInjectorOption
	RejectInjectorOption
	ErrorInjectorOption
	SlowInjectorOption
}

type errorOptionBool bool

func (o errorOptionBool) applyFault(f *Fault) error {
	return errErrorOption
}

func (o errorOptionBool) applyChainInjector(f *ChainInjector) error {
	return errErrorOption
}

func (o errorOptionBool) applyRandomInjector(f *RandomInjector) error {
	return errErrorOption
}

func (o errorOptionBool) applyRejectInjector(f *RejectInjector) error {
	return errErrorOption
}

func (o errorOptionBool) applyErrorInjector(f *ErrorInjector) error {
	return errErrorOption
}

func (o errorOptionBool) applySlowInjector(f *SlowInjector) error {
	return errErrorOption
}

func withError() errorOption {
	return errorOptionBool(true)
}

// testReporter is a reporter that records states for assertions.
type testReporter struct {
	t      *testing.T
	mu     sync.Mutex
	states []InjectorState
}

// newTestReporter returns a new testReporter.
func newTestReporter(t *testing.T) *testReporter {
	return &testReporter{t: t}
}

// Report records the state.
func (r *testReporter) Report(name string, state InjectorState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states = append(r.states, state)
}

// hasState returns true if the given state was reported.
func (r *testReporter) hasState(state InjectorState) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.states {
		if s == state {
			return true
		}
	}
	return false
}
