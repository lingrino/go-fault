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

var benchmarkHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, benchmarkHandlerBody, benchmarkHandlerCode)
})

func sendRequestBenchmark(b *testing.B, f *fault.Fault) *httptest.ResponseRecorder {
	b.Helper()

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

func runBenchmark(b *testing.B, f *fault.Fault) {
	var rr *httptest.ResponseRecorder

	for n := 0; n < b.N; n++ {
		rr = sendRequestBenchmark(b, f)
	}

	result = rr
}

func BenchmarkNoFault(b *testing.B) {
	runBenchmark(b, nil)
}

func BenchmarkFaultDisabled(b *testing.B) {
	i, _ := fault.NewErrorInjector(500)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           false,
		Injector:          i,
		PercentOfRequests: 0.0,
	})

	runBenchmark(b, f)
}

func BenchmarkFaultErrorZeroPercent(b *testing.B) {
	i, _ := fault.NewErrorInjector(500)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          i,
		PercentOfRequests: 0.0,
	})

	runBenchmark(b, f)
}

func BenchmarkFaultError100Percent(b *testing.B) {
	i, _ := fault.NewErrorInjector(500)

	f, _ := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          i,
		PercentOfRequests: 1.0,
	})

	runBenchmark(b, f)
}
