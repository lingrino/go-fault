package fault_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/github/go-fault"
)

var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "OK", http.StatusOK)
})

func Example() {
	// Wait one millisecond then continue
	si, err := fault.NewSlowInjector(time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	// Return a 500
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	if err != nil {
		log.Fatal(err)
	}

	// Chain slow and error injectors together
	ci, err := fault.NewChainInjector(si, ei)
	if err != nil {
		log.Fatal(err)
	}

	// Run our fault injection 100% of the time
	f, err := fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          ci,
		PercentOfRequests: 1.0,
		PathBlacklist:     []string{"/ping", "/health"},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Insert our middleware before the mainHandler
	handlerChain := f.Handler((mainHandler))

	// Create a dummy request and response records
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Run our request
	handlerChain.ServeHTTP(rr, req)

	// Verify the correct response
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())
	// Output:
	// 500
	// Internal Server Error
}
