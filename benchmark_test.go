package fault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/github/go-fault"
)

// benchmarkRequest simulates a request with the provided Fault injected.
func benchmarkRequest(b *testing.B, f *fault.Fault) *httptest.ResponseRecorder {
	b.Helper()

	// benchmarkHandler is the main handler that runs on our request.
	var benchmarkHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "OK", http.StatusOK)
	})

	// If we instead use httptest.NewRequest here our benchmark times will approximately double.
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	if f != nil {
		finalHandler := f.Handler(benchmarkHandler)
		finalHandler.ServeHTTP(rr, req)
	} else {
		benchmarkHandler.ServeHTTP(rr, req)
	}

	return rr
}

// runBenchmark runs the benchmark look with the provided Fault.
func runBenchmark(b *testing.B, f *fault.Fault) {
	var rr *httptest.ResponseRecorder

	for n := 0; n < b.N; n++ {
		rr = benchmarkRequest(b, f)
	}

	_ = rr
}

// BenchmarkNoFault is our control with no Fault injected.
func BenchmarkNoFault(b *testing.B) {
	runBenchmark(b, nil)
}

// BenchmarkFaultDisabled benchmarks a disabled Fault struct.
func BenchmarkFaultDisabled(b *testing.B) {
	i, _ := fault.NewErrorInjector(http.StatusInternalServerError)
	f, _ := fault.NewFault(fault.Options{
		Enabled:           false,
		Injector:          i,
		PercentOfRequests: 0.0,
	})

	runBenchmark(b, f)
}

// BenchmarkFaultErrorZeroPercent benchmarks an enabled fault struct that runs 0% of the time.
func BenchmarkFaultErrorZeroPercent(b *testing.B) {
	i, _ := fault.NewErrorInjector(http.StatusInternalServerError)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          i,
		PercentOfRequests: 0.0,
	})

	runBenchmark(b, f)
}

// BenchmarkFaultError100Percent benchmarks an enabled fault struct that runs 100% of the time.
func BenchmarkFaultError100Percent(b *testing.B) {
	i, _ := fault.NewErrorInjector(http.StatusInternalServerError)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          i,
		PercentOfRequests: 1.0,
	})

	runBenchmark(b, f)
}
