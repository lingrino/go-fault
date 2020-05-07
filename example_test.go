package fault_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/github/go-fault"
)

// ExampleNewFault shows how to create a new Fault.
func ExampleNewFault() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
	)

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewFault_blocklist shows how to create a new Fault with a path blocklist.
func ExampleNewFault_blocklist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
		fault.WithPathBlocklist([]string{"/ping", "/health"}),
	)

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewFault_allowlist shows how to create a new Fault with a path allowlist.
func ExampleNewFault_allowlist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
		fault.WithPathAllowlist([]string{"/injecthere", "/andhere"}),
	)

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewChainInjector shows how to create a new ChainInjector.
func ExampleNewChainInjector() {
	si, err := fault.NewSlowInjector(time.Minute)
	ri, err := fault.NewRejectInjector()

	_, err = fault.NewChainInjector([]fault.Injector{si, ri})

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewChainInjector shows how to create a new RandomInjector.
func ExampleNewRandomInjector() {
	si, err := fault.NewSlowInjector(time.Minute)
	ri, err := fault.NewRejectInjector()

	_, err = fault.NewRandomInjector([]fault.Injector{si, ri})

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewRejectInjector shows how to create a new RejectInjector.
func ExampleNewRejectInjector() {
	_, err := fault.NewRejectInjector()

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewErrorInjector shows how to create a new ErrorInjector.
func ExampleNewErrorInjector() {
	_, err := fault.NewErrorInjector(http.StatusInternalServerError)

	fmt.Println(err)
	// Output: <nil>
}

// ExampleNewSlowInjector shows how to create a new SlowInjector.
func ExampleNewSlowInjector() {
	_, err := fault.NewSlowInjector(time.Second * 10)

	fmt.Println(err)
	// Output: <nil>
}
