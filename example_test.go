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

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
	)

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewFault_blacklist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
		fault.WithPathBlacklist([]string{"/ping", "/health"}),
	)

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewFault_whitelist() {
	ei, err := fault.NewErrorInjector(http.StatusInternalServerError)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fault.NewFault(ei,
		fault.WithEnabled(true),
		fault.WithParticipation(0.25),
		fault.WithPathWhitelist([]string{"/injecthere", "/andhere"}),
	)

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

	_, err = fault.NewChainInjector([]fault.Injector{si, ri})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewRandomInjector() {
	si, err := fault.NewSlowInjector(time.Minute)
	if err != nil {
		log.Fatal(err)
	}

	ri, err := fault.NewRejectInjector()
	if err != nil {
		log.Fatal(err)
	}

	_, err = fault.NewRandomInjector([]fault.Injector{si, ri})
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
