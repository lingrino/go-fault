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

// runBenchmark benchmarks the provided Fault.
func runBenchmark(b *testing.B, f *fault.Fault) {
	var rr *httptest.ResponseRecorder

	for n := 0; n < b.N; n++ {
		rr = benchmarkRequest(b, f)
	}

	_ = rr
}

// BenchmarkNoFault is our control using no Fault.
func BenchmarkNoFault(b *testing.B) {
	runBenchmark(b, nil)
}

// BenchmarkFaultDisabled benchmarks a disabled Fault.
func BenchmarkFaultDisabled(b *testing.B) {
	i, _ := fault.NewErrorInjector(http.StatusInternalServerError)
	f, _ := fault.NewFault(i,
		fault.WithEnabled(false),
	)

	runBenchmark(b, f)
}

// BenchmarkFaultErrorZeroPercent benchmarks an enabled Fault with 0% participation.
func BenchmarkFaultErrorZeroPercent(b *testing.B) {
	i, _ := fault.NewErrorInjector(http.StatusInternalServerError)
	f, _ := fault.NewFault(i,
		fault.WithEnabled(true),
		fault.WithParticipation(0.0),
	)

	runBenchmark(b, f)
}

// BenchmarkFaultError100Percent benchmarks an enabled Fault with 100% participation.
func BenchmarkFaultError100Percent(b *testing.B) {
	i, _ := fault.NewErrorInjector(http.StatusInternalServerError)
	f, _ := fault.NewFault(i,
		fault.WithEnabled(true),
		fault.WithParticipation(1.0),
	)

	runBenchmark(b, f)
}
