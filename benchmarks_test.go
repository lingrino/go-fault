package fault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/github/go-fault"
)

var result *httptest.ResponseRecorder

const (
	benchmarkHandlerCode = http.StatusOK
	benchmarkHandlerBody = "OK"
)

// benchmarkHandler simulates a good request. When no faults are enabled we should
// expect this result back immediately.
var benchmarkHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, benchmarkHandlerBody, benchmarkHandlerCode)
})

// sendRequestBenchmark abstracts sending a standard request with N number
// of faults chained before our benchmarkHandler. The faults that are passed
// first in the list will execute last in the chain.
func sendRequestBenchmark(b *testing.B, fs ...*fault.Fault) *httptest.ResponseRecorder {
	b.Helper()

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	if fs != nil {
		finalHandler := fs[0].Handler(benchmarkHandler)
		for _, f := range fs[1:] {
			finalHandler = f.Handler(finalHandler)
		}
		finalHandler.ServeHTTP(rr, req)
	} else {
		benchmarkHandler.ServeHTTP(rr, req)
	}

	return rr
}

// runBenchmark abstracts benchmarking the request
func runBenchmark(b *testing.B, fs ...*fault.Fault) {
	var rr *httptest.ResponseRecorder

	for n := 0; n < b.N; n++ {
		rr = sendRequestBenchmark(b, fs...)
	}

	result = rr
}

// BenchmarkNoFault is our baseline benchmark to compare others against.
// It benchmarks requests against our benchmarkHandler without a fault
func BenchmarkNoFault(b *testing.B) {
	runBenchmark(b)
}

// BenchmarkFaultDisabled benchmarks with the fault middleware in the
// request path but disabled.
func BenchmarkFaultDisabled(b *testing.B) {
	i, _ := fault.NewErrorInjector(500)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           false,
		Injector:          i,
		PercentOfRequests: 0.0,
	})

	runBenchmark(b, f)
}

// BenchmarkFaultErrorZeroPercent benchmarks the fault.Error Type when
// the fault is enabled but PercentOfRequests is 0.0
func BenchmarkFaultErrorZeroPercent(b *testing.B) {
	i, _ := fault.NewErrorInjector(500)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          i,
		PercentOfRequests: 0.0,
	})

	runBenchmark(b, f)
}

// BenchmarkFaultError100Percent benchmarks the fault.Error Type when
// the fault is enabled and PercentOfRequests is 1.0
func BenchmarkFaultError100Percent(b *testing.B) {
	i, _ := fault.NewErrorInjector(500)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          i,
		PercentOfRequests: 1.0,
	})

	runBenchmark(b, f)
}
