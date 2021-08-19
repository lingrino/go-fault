# Fault

[![PkgGoDev](https://pkg.go.dev/badge/github.com/github/go-fault)](https://pkg.go.dev/github.com/github/go-fault) [![goreportcard](https://goreportcard.com/badge/github.com/github/go-fault)](https://goreportcard.com/report/github.com/github/go-fault)

The fault package provides go http middleware that makes it easy to inject faults into your service. Use the fault package to reject incoming requests, respond with an HTTP error, inject latency into a percentage of your requests, or inject any of your own custom faults.

## Features

The fault package works through [standard go http middleware](https://pkg.go.dev/net/http/?tab=doc#Handler). You first create an `Injector`, which is a middleware with the code to be run on injection. Then you wrap that `Injector` in a `Fault` which handles logic about when to run your `Injector`.

There are currently three kinds of injectors: `SlowInjector`, `ErrorInjector`, and `RejectInjector`. Each of these injectors can be configured through a `Fault` to run on a small percent of your requests. You can also configure the `Fault` to blocklist/allowlist certain paths.

See the usage section below for an example of how to get started and the [godoc](https://pkg.go.dev/github.com/github/go-fault?tab=doc) for further documentation.

## Limitations

This package is useful for safely testing failure scenarios in go services that can make use of `net/http` handlers/middleware.

One common failure scenario that we cannot perfectly simulate is dropped requests. The `RejectInjector` will always return immediately to the user, but in many cases requests can be dropped without ever sending a response. The best way to simulate this scenario using the fault package is to chain a `SlowInjector` with a very long wait time in front of an eventual `RejectInjector`.

## Status

This project is in a stable and supported state. There are no plans to introduce significant new features however we welcome and encourage any ideas and contributions from the community. Contributions should follow the guidelines in our [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## Usage

```go
// main.go
package main

import (
        "net/http"
        "time"

        "github.com/github/go-fault"
)

var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
})

func main() {
        slowInjector, _ := fault.NewSlowInjector(time.Second * 2)
        slowFault, _ := fault.NewFault(slowInjector,
                fault.WithEnabled(true),
                fault.WithParticipation(0.25),
                fault.WithPathBlocklist([]string{"/ping", "/health"}),
        )

        // Add 2 seconds of latency to 25% of our requests
        handlerChain := slowFault.Handler(mainHandler)

        http.ListenAndServe("127.0.0.1:3000", handlerChain)
}
```

## Development

This package uses standard go tooling for testing and development. The [go](https://golang.org/dl/) language is all you need to contribute. Tests use the popular [testify/assert](https://github.com/stretchr/testify/) which will be downloaded automatically the first time you run tests. GitHub Actions will also run a linter using [golangci-lint](https://github.com/golangci/golangci-lint) after you push. You can also download the linter and use `golangci-lint run` to run locally.

## Testing

The fault package has extensive tests that are run in [GitHub Actions](https://github.com/github/go-fault/actions?query=workflow%3AValidate) on every push. Code coverage is 100% and is published as an artifact on every Actions run.

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

## Maintainers

[@lingrino](https://github.com/lingrino)

### Contributors

[@mrfaizal](https://github.com/mrfaizal)
[@vroldanbet](https://github.com/vroldanbet)
[@fatih](https://github.com/fatih)

## License

This project is licensed under the [MIT License](LICENSE.md).

## Want to file an issue?

Please head over to the [github.com/github/go](https://github.com/github/go/issues) repository. The frameworks-services team owns many libraries, and keeping issues filed in one place helps us track and prioritize them accordingly. Thank you!
