# Fault

[GoDoc](https://godoc.githubapp.com/github.com/github/go-fault)

The fault package provides go middleware that makes it easy to inject faults into your service. Use the fault package to reject incoming requests, respond with an HTTP error, inject latency into a percentage of your requests, or evaluate any of your own custom faults.

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

The fault package has extensive tests that are run in [GitHub Actions](https://github.com/github/go-fault/actions?query=workflow%3AValidate) on every push.

You can also run tests locally:

```shell
$ go test -v -cover -race ./...
[...]
PASS
coverage: 100.0% of statements
ok      github.com/github/go-fault      0.575s
```

## Benchmarks

The fault package is safe to leave implemented even when you are not running a fault injection. While the fault is disabled there is negligible performance degradation compared to removing the package from the request path. While enabled there may be minor performance differences, but this will only be the case *while you are already injecting faults.*

Benchmarks are provided to compare without faults, with faults disabled, and with faults enabled. Benchmarks are uploaded as artifacts in GitHub Actions and you can download them from any [Validate Workflow](https://github.com/github/go-fault/actions?query=workflow%3AValidate).

You can also run benchmarks locally (example output):

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
