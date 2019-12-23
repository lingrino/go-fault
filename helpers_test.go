package fault_test

import (
	"github.com/github/fault"
	"net/http"
	"net/http/httptest"
	"testing"
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

func testRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	return req
}

func sendRequest(t *testing.T, f *fault.Fault) *httptest.ResponseRecorder {
	t.Helper()

	req := testRequest(t)
	rr := httptest.NewRecorder()
	app := f.Handler(testHandler)
	app.ServeHTTP(rr, req)

	return rr
}
