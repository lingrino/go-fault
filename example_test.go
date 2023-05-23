package fault_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lingrino/go-fault"
)

// ExampleNewFault shows how to create a new Fault.
func ExampleNewFault() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	fmt.Print(err)

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
	)

	fmt.Print(err)
	// Output: <nil><nil>
}

// ExampleNewFault_blocklist shows how to create a new Fault with a path/header blocklist.
func ExampleNewFault_blocklist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	fmt.Print(err)

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
		fault.WithPathBlocklist([]string{"/ping", "/health"}),
		fault.WithHeaderBlocklist(map[string]string{"block": "this header"}),
	)

	fmt.Print(err)
	// Output: <nil><nil>
}

// ExampleNewFault_allowlist shows how to create a new Fault with a path/header allowlist.
func ExampleNewFault_allowlist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	fmt.Print(err)

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
		fault.WithPathAllowlist([]string{"/injecthere", "/andhere"}),
		fault.WithHeaderAllowlist(map[string]string{"allow": "this header"}),
	)

	fmt.Print(err)
	// Output: <nil><nil>
}

// ExampleNewChainInjector shows how to create a new ChainInjector.
func ExampleNewChainInjector() {
	si, err := fault.NewSlowInjector(time.Minute)
	fmt.Print(err)
	ri, err := fault.NewRejectInjector()
	fmt.Print(err)

	_, err = fault.NewChainInjector([]fault.Injector{si, ri})

	fmt.Print(err)
	// Output: <nil><nil><nil>
}

// ExampleNewChainInjector shows how to create a new RandomInjector.
func ExampleNewRandomInjector() {
	si, err := fault.NewSlowInjector(time.Minute)
	fmt.Print(err)
	ri, err := fault.NewRejectInjector()
	fmt.Print(err)

	_, err = fault.NewRandomInjector([]fault.Injector{si, ri})

	fmt.Print(err)
	// Output: <nil><nil><nil>
}

// ExampleNewRejectInjector shows how to create a new RejectInjector.
func ExampleNewRejectInjector() {
	_, err := fault.NewRejectInjector()

	fmt.Print(err)
	// Output: <nil>
}

// ExampleNewErrorInjector shows how to create a new ErrorInjector.
func ExampleNewErrorInjector() {
	_, err := fault.NewErrorInjector(http.StatusInternalServerError)

	fmt.Print(err)
	// Output: <nil>
}

// ExampleNewSlowInjector shows how to create a new SlowInjector.
func ExampleNewSlowInjector() {
	_, err := fault.NewSlowInjector(time.Second * 10)

	fmt.Print(err)
	// Output: <nil>
}
