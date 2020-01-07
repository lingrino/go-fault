package fault_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/github/fault"
)

const (
	testHandlerCode        = http.StatusOK
	testHandlerContentType = "application/json"
	testHandlerBody        = `{"status": "OK"}`
)

// testHandler simulates a good request. When no faults are enabled we should
// expect this result back immediately.
var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(testHandlerCode)
	w.Header().Set("Content-Type", testHandlerContentType)
	w.Write([]byte(testHandlerBody))
})

// testRequest abstracts creating a standard request that we use in all tests
func testRequest(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	return req
}

// sendRequest abstracts sending a standard request with N number of faults
// chained before our testHandler. The faults that are passed first in the
// list will execute last in the chain.
func sendRequest(t *testing.T, fs ...*fault.Fault) *httptest.ResponseRecorder {
	t.Helper()

	req := testRequest(t, context.Background())
	rr := httptest.NewRecorder()

	app := fs[0].Handler(testHandler)

	for _, f := range fs[1:] {
		app = f.Handler(app)
	}

	app.ServeHTTP(rr, req)

	return rr
}

// sendRequestExpectTimeout does the same as sendRequest except that the test will
// fail if we don't receive a timeout in the configured amount of time.
func sendRequestExpectTimeout(t *testing.T, f *fault.Fault, to time.Duration) *httptest.ResponseRecorder {
	t.Helper()

	done := make(chan bool, 1)

	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	req := testRequest(t, ctx)
	rr := httptest.NewRecorder()
	app := f.Handler(testHandler)

	go func() {
		// If we don't reach timeout it's common in our tests that we panic, catch that here
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected: fail with timeout %v got: panic: %v", to, r)
			}
		}()
		app.ServeHTTP(rr, req)
		done <- true
	}()

	select {
	case <-done:
		t.Errorf("expected to fail with timeout %v", to)
	case <-ctx.Done():
		return rr
	}

	return rr
}

// sendRequestExpectPanic does the same as sendRequest except that the test
// will fail if we don't receive an http.ErrAbortHandler panic.
func sendRequestExpectPanic(t *testing.T, f *fault.Fault) *httptest.ResponseRecorder {
	t.Helper()

	// Recover from our expected http.ErrAbortHandler panics but fail on others
	defer func() {
		if r := recover(); r != nil {
			if r != http.ErrAbortHandler {
				t.Fatal(r)
			}
		}
	}()

	req := testRequest(t, context.Background())
	rr := httptest.NewRecorder()
	app := f.Handler(testHandler)
	app.ServeHTTP(rr, req)

	return rr
}
