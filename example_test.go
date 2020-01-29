package fault_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/github/go-fault"
)

func ExampleNewFault() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	if err != nil {
		log.Fatal(err)
	}

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
	if err != nil {
		log.Fatal(err)
	}

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
	if err != nil {
		log.Fatal(err)
	}

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
	if err != nil {
		log.Fatal(err)
	}

	ri, err := fault.NewRejectInjector()
	if err != nil {
		log.Fatal(err)
	}

	_, err = fault.NewChainInjector(si, ri)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewRejectInjector() {
	_, err := fault.NewRejectInjector()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewErrorInjector() {
	_, err := fault.NewErrorInjector(http.StatusInternalServerError)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewSlowInjector() {
	_, err := fault.NewSlowInjector(time.Second * 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(err)
	// Output: <nil>
}
