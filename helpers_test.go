package fault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	// We don't use http.StatusOK because some operations default to that and then we can't tell
	// the difference between what testHandler wrote and what the operation wrote.
	testHandlerCode = http.StatusAccepted
	testHandlerBody = "Accepted"
)

// testRequest simulates a request with the provided Fault injected.
func testRequest(t *testing.T, f *Fault) *httptest.ResponseRecorder {
	t.Helper()

	// testHandler is the main handler that runs on our request.
	var testHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, testHandlerBody, testHandlerCode)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	if f != nil {
		finalHandler := f.Handler(testHandler)
		finalHandler.ServeHTTP(rr, req)
	} else {
		testHandler.ServeHTTP(rr, req)
	}

	return rr
}

// testRequestExpectPanic runs testRequest and catches panics, failing the test if the panic is not
// http.ErrAbortHandler.
func testRequestExpectPanic(t *testing.T, f *Fault) *httptest.ResponseRecorder {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			if r != http.ErrAbortHandler {
				t.Fatal(r)
			}
		}
	}()

	rr := testRequest(t, f)

	return rr
}

// testInjector is a simple Injector used for running tests. By default testInjector just passes on
// the request but if you set resp500 to true then it will instead return a 500.
type testInjector struct {
	resp500 bool
}

// newTestInjector creates a new testInjector struct.
func newTestInjector(resp500 bool) *testInjector {
	return &testInjector{
		resp500: resp500,
	}
}

// Handler returns a 500 if resp500 is true and otherwise passes on the request.
func (i *testInjector) Handler(next http.Handler) http.Handler {
	if i.resp500 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(500), 500)
		})
	}
	return next
}
