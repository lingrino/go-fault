package fault_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/github/go-fault"
)

// Example is a package-level documentation example.
func Example() {
	// Wait one millisecond then continue
	si, _ := fault.NewSlowInjector(time.Millisecond)

	// Return a 500
	ei, _ := fault.NewErrorInjector(http.StatusInternalServerError)

	// Chain slow and error injectors together
	ci, _ := fault.NewChainInjector([]fault.Injector{si, ei})

	// Run our fault injection 100% of the time
	f, _ := fault.NewFault(ci,
		fault.WithEnabled(true),
		fault.WithParticipation(1.0),
		fault.WithPathBlacklist([]string{"/ping", "/health"}),
	)

	// mainHandler responds 200/OK
	var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "OK", http.StatusOK)
	})

	// Insert our middleware before the mainHandler
	handlerChain := f.Handler((mainHandler))

	// Create dummy request and response records
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Run our request
	handlerChain.ServeHTTP(rr, req)

	// Verify the correct response
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())
	// Output: 500
	// Internal Server Error
}
