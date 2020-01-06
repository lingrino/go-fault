# Fault

The fault package is a go middleware that makes it easy to inject faults into your service. Use the fault package to reject incoming requests, respond with an HTTP error, or inject latency into a percentage of your requests.

## Documentation

For detailed package documentation you can run `go doc` or run a godoc server locally by running `godoc -http=:6060` and then visiting <localhost:6060/pkg/github.com/github/fault/> (make sure you cloned into your $GOPATH) in your browser.

## Usage

```go
// main.go
package main

import (
        "net/http"

        "github.com/github/fault"
)

var mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
})

func main() {
        slowFault := fault.New(fault.Options{
                Enabled:           true,
                Type:              fault.TypeSlow,
                Value:             2000, // 2 seconds
                PercentOfRequests: 0.25, // 25%
        })

        handlerChain := slowFault.Handler(mainHandler)

        http.ListenAndServe("127.0.0.1:3000", handlerChain)
}
```

## Testing Locally

This repo contains a Docker image and simple implementation of the fault package in the [test](./test/main.go) directory. Use the docker image to see how the fault package behaves by building and running requests against it locally.

```shell
$ docker build -t fault .
Successfully built d41114b355ee
$ docker run -d -p 3000:3000 fault
c6c6f3a03f1eb79158deb1fbb3dabd36488d5ef9290efc213419e257b00ce9e6
$ curl -v localhost:3000
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 3000 (#0)
> GET / HTTP/1.1
> Host: localhost:3000
> User-Agent: curl/7.64.1
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< Date: Mon, 06 Jan 2020 20:44:24 GMT
< Content-Length: 0
<
* Connection #0 to host localhost left intact
* Closing connection 0
```

## Testing

The fault package has extensive tests.

```shell
$ go test -v -cover -race
coverage: 100.0% of statements
ok      github.com/github/fault 2.970s
```

## Benchmarks

The fault package is safe to leave implemented even when you are not running a fault injection. If you set `fault.Opt.Enabled` to `false` there should be no performance degradation compared to removing the package from the request path. If you have `fault.Opt.Enabled` set to `true` there may be very minor performance differences, but this will only be the case *while you are already doing fault injection.*

There will soon be benchmarks that verify these claims and prevent changes that cause performance degradation.

## Status

The package is mostly implemented, however breaking API changes may still happen before the `v1.0.0` release. The fault package is intentionally simple and new features are unlikely to be implemented. Here are a few things that may still be added before `v1.0.0`.

- Option to make chained middleware depend on earlier faults (% slow then always error)
- Option to set a context value on the request if fault injection was done
- Option to return a header (will not work on reject) if fault injection was done
- Option to customize the response body on error faults
- Option to slow requests in a random distribution instead of a fixed value
- Option to always run faults if a certain header is passed
