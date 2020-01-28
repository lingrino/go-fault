package fault_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/github/go-fault"
)

func ExampleNewFault() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)

	_, err = fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          ei,
		PercentOfRequests: 1.0,
	})

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewFault_blacklist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)

	_, err = fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          ei,
		PercentOfRequests: 1.0,
		PathBlacklist:     []string{"/ping", "/health"},
	})

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewFault_whitelist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)

	_, err = fault.NewFault(fault.Options{
		Enabled:           true,
		Injector:          ei,
		PercentOfRequests: 1.0,
		PathWhitelist:     []string{"/injecthere", "/andhere"},
	})

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewChainInjector() {
	si, err := fault.NewSlowInjector(time.Minute)
	ri, err := fault.NewRejectInjector()
	_, err = fault.NewChainInjector(si, ri)

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewRejectInjector() {
	_, err := fault.NewRejectInjector()

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewErrorInjector() {
	_, err := fault.NewErrorInjector(http.StatusInternalServerError)

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewSlowInjector() {
	_, err := fault.NewSlowInjector(10 * time.Second)

	fmt.Println(err)
	// Output: <nil>
}
