package fault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	// Don't use http.StatusOK because some operations default to that and then we can't tell
	// the difference between what testHandler wrote and what the operation wrote
	testHandlerCode = http.StatusAccepted
	testHandlerBody = "Accepted"
)

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	http.Error(w, testHandlerBody, testHandlerCode)
})

func testRequest(t *testing.T, fs ...*Fault) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	finalHandler := fs[0].Handler(testHandler)
	for _, f := range fs[1:] {
		finalHandler = f.Handler(finalHandler)
	}

	finalHandler.ServeHTTP(rr, req)

	return rr
}

func testRequestExpectPanic(t *testing.T, fs ...*Fault) *httptest.ResponseRecorder {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			if r != http.ErrAbortHandler {
				t.Fatal(r)
			}
		}
	}()

	rr := testRequest(t, fs...)

	return rr
}

type testInjector struct {
	resp500 bool
}

func newTestInjector(resp500 bool) *testInjector {
	return &testInjector{
		resp500: resp500,
	}
}

func (i *testInjector) Handler(next http.Handler) http.Handler {
	if i.resp500 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(500), 500)
			return
		})
	}
	return next
}
