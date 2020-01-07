package fault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/github/fault"
)

var result *httptest.ResponseRecorder

const (
	benchmarkHandlerCode        = http.StatusOK
	benchmarkHandlerContentType = "application/json"
	benchmarkHandlerBody        = `{"status": "OK"}`
)

// benchmarkHandler simulates a good request. When no faults are enabled we should
// expect this result back immediately.
var benchmarkHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(benchmarkHandlerCode)
	w.Header().Set("Content-Type", benchmarkHandlerContentType)
	w.Write([]byte(benchmarkHandlerBody))
})

// benchmarkRequest abstracts creating a standard request that we use in all benchmarks
func benchmarkRequest(b *testing.B) *http.Request {
	b.Helper()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		b.Fatal(err)
	}

	return req
}

// sendRequestBenchmark abstracts sending a standard request with N number
// of faults chained before our benchmarkHandler. The faults that are passed
// first in the list will execute last in the chain.
func sendRequestBenchmark(b *testing.B, fs ...*fault.Fault) *httptest.ResponseRecorder {
	b.Helper()

	req := benchmarkRequest(b)
	rr := httptest.NewRecorder()

	if fs != nil {
		app := fs[0].Handler(benchmarkHandler)

		for _, f := range fs[1:] {
			app = f.Handler(app)
		}

		app.ServeHTTP(rr, req)
	} else {
		benchmarkHandler.ServeHTTP(rr, req)
	}

	return rr
}

// BenchmarkNoFault is our baseline benchmark to compare others against.
// It benchmarks requests against our benchmarkHandler without a fault
func BenchmarkNoFault(b *testing.B) {
	var rr *httptest.ResponseRecorder

	for n := 0; n < b.N; n++ {
		rr = sendRequestBenchmark(b)
	}

	result = rr
}

// BenchmarkFaultDisabled benchmarks with the fault middleware in the
// request path but disabled.
func BenchmarkFaultDisabled(b *testing.B) {
	var rr *httptest.ResponseRecorder

	f := fault.New(fault.Options{
		Enabled: false,
	})

	for n := 0; n < b.N; n++ {
		rr = sendRequestBenchmark(b, f)
	}

	result = rr
}

// BenchmarkFaultErrorZeroPercent benchmarks the fault.Error Type when
// the fault is enabled but PercentOfRequests is 0.0
func BenchmarkFaultErrorZeroPercent(b *testing.B) {
	var rr *httptest.ResponseRecorder

	f := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.Error,
		Value:             500,
		PercentOfRequests: 0.0,
	})

	for n := 0; n < b.N; n++ {
		rr = sendRequestBenchmark(b, f)
	}

	result = rr
}

// BenchmarkFaultError100Percent benchmarks the fault.Error Type when
// the fault is enabled and PercentOfRequests is 1.0
func BenchmarkFaultError100Percent(b *testing.B) {
	var rr *httptest.ResponseRecorder

	f := fault.New(fault.Options{
		Enabled:           true,
		Type:              fault.Error,
		Value:             500,
		PercentOfRequests: 1.0,
	})

	for n := 0; n < b.N; n++ {
		rr = sendRequestBenchmark(b, f)
	}

	result = rr
}
