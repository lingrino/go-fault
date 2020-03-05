# Fault

The fault package provides go middleware that makes it easy to inject faults into your service. Use the fault package to reject incoming requests, respond with an HTTP error, inject latency into a percentage of your requests, or evaluate any of your own custom faults.

## Documentation



For detailed package documentation and examples you can run `go doc` or run a godoc server locally by running `godoc -http=:6060` and then visiting <http://localhost:6060/pkg/github.com/github/go-fault/> (make sure you cloned into your $GOPATH) in your browser.

## Usage

```go
// main.go
package main

import (
        "net/http"

        "github.com/github/go-fault"
)

var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, testHandlerBody, testHandlerCode)
})

func main() {
        slowInjector, _ := fault.NewSlowInjector(time.Second * 2)
        slowFault, _ := fault.NewFault(slowInjector,
                fault.WithEnabled(true),
                fault.WithParticipation(0.25),
                fault.WithPathBlacklist([]string{"/ping", "/health"}),
        )

        // Add 2 seconds of latency to 25% of our requests
        handlerChain := slowFault.Handler(mainHandler)

        http.ListenAndServe("127.0.0.1:3000", handlerChain)
}
```

## Testing

The fault package has extensive tests.

```shell
$ go test -v -cover -race
coverage: 100.0% of statements
ok      github.com/github/go-fault      2.858s
```

## Benchmarks

The fault package is safe to leave implemented even when you are not running a fault injection. While the fault is disabled there should be negligible performance degradation compared to removing the package from the request path. While enabled there may be very minor performance differences, but this will only be the case *while you are already doing fault injection.*

Benchmarks are provided to compare without faults, with faults disabled, and with faults enabled. Benchmarks are uploaded as artifacts in the GitHub Actions Validate Workflow and you can download them from any [Validate Workflow](https://github.com/github/go-fault/actions?query=workflow%3AValidate).

You can also run benchmarks locally using the following command (example output):

```shell
$ go test -run=XXX -bench=.
goos: darwin
goarch: amd64
pkg: github.com/github/go-fault
BenchmarkNoFault-8                        684826              1734 ns/op
BenchmarkFaultDisabled-8                  675291              1771 ns/op
BenchmarkFaultErrorZeroPercent-8          667903              1823 ns/op
BenchmarkFaultError100Percent-8           663661              1833 ns/op
PASS
ok      github.com/github/go-fault      8.814s
```

## Status

The package is mostly implemented, however breaking API changes may still happen before the `v1.0.0` release. The fault package is intentionally simple and new features are unlikely to be implemented. Here are a few things that may still be added before `v1.0.0`.

- Option to always run faults if a certain header is passed
