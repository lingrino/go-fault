/*
Package fault provides standard http middleware for fault injection in go

Basics

Use the fault package to inject faults into the http request path of your service.
Fault works by modfying and/or delaying your service's http responses. Place the
middleware high enough in the chain that it can act quickly, but after any other
middlewares that should complete before fault injection (auth, redirects, etc...)

The type and severity of injected faults is controlled by a single Options struct
passed to a new fault handler. The package provides no other means of configuring
the faults and it is up to the consumer of the package to decide how to manage the
fault options (feature flags, environment variables, deploys, etc...).

Note that you can always insert the middleware twice if you (for example) want
to REJECT a request after a 100ms SLOW.

Fault will always default to a "do nothing, pass request on" state if the provided
options are invalid. If you are not seeing faults injected like you expect you may
have passed an out of bounds value, invalid http status code, incorrect percent, or
other wrong parameter.

Fault Types

There are three type of faults that can be injected.

	fault.TypeReject
	fault.TypeError
	fault.TypeSlow

Reject

Use fault.TypeReject to immediately return an empty response. For example, a curl
for a rejected response will produce:

	$ curl https://github.com
	curl: (52) Empty reply from server

Error

Use fault.TypeError to immediately return an http status code of your choosing.
For example, you can return a 200, 301, 404, 500, or any other valid status code
to test how your clients respond to different statuses. If you enter an invalid
status code the middleware will pass on the request without a response.

The Error fault type works by writing the provided status code to the response
and then returning immediately. The response will preserve any non-status parts
of the header or body that you have already written.

Slow

Use fault.TypeSlow to wait a configured amount of milliseconds before proceeding
with the request as normal. For example, you can use the Slow fault type to add
a 10ms delay to half of your incoming requests.

Combining Faults

It is easy to combine any of the fault types into a chained action. There are two
ways you might want to combine faults.

First, you can have fault types that are sequential but independent of each other.
For example, you can chain them such that 1% of requests will return a 500 error
and 1% of requests will be rejected.

Second, you might want to combine faults such that 1% of requests will be slowed
for 10ms and then rejected. These two faults depend on each other. It is possible
to add this capability with the fault package by setting the TODO flag in the
chained (2nd) Options struct. When you do this the second fault will only be
injected if the fault before it was activated.

Configuration

All configuration for the fault package is done through the Options struct. There
is no other way to manage configuration for the package. It is up to the user of
the fault package to manage how the Options struct is generated. Common options
are feature flags, environment variables, or code changes in deploys.

*/
package fault
