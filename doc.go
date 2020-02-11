/*
Package fault provides standard http middleware for fault injection in go.

Basics

Use the fault package to inject faults into the http request path of your service. Faults work by
modifying and/or delaying your service's http responses. Place the Fault middleware high enough in
the chain that it can act quickly, but after any other middlewares that should complete before fault
injection (auth, redirects, etc...).

The type and severity of injected faults is controlled by a single Options struct passed to a new
Fault struct. The Options struct must contain a field Injector, which is an interface that holds the
actual fault injection code in Injector.Handler. The Fault struct wraps Injector.Handler in another
Fault.Handler that applies generic Fault logic (such as what % of requests to run the Injector on)
to the Injector.

Package provided Handlers will always default to a "do nothing, pass request on" state if the
provided options are invalid. Make sure you use the NewFault() and NewTypeInjector() constructors to
create validated Faults and Injectors. If you are not seeing faults injected like you expect you may
have passed an out of bounds value, invalid http status code, incorrect percent, or other wrong
parameter.

Injectors

There are three main Injectors provided by the fault package:

    fault.RejectInjector
    fault.ErrorInjector
    fault.SlowInjector

RejectInjector

Use fault.RejectInjector to immediately return an empty response. For example, a curl for a rejected
response will produce:

    $ curl https://github.com
    curl: (52) Empty reply from server

ErrorInjector

Use fault.ErrorInjector to immediately return an http status code of your choosing along with the
standard HTTP response body for that code. For example, you can return a 200, 301, 418, 500, or any
other valid status code to test how your clients respond to different statuses. If ErrorInjector has
an invalid status code the middleware will pass on the request without evaluating.

SlowInjector

Use fault.SlowInjector to wait a configured time.Duration before proceeding with the request as
normal. For example, you can use the SlowInjector to add a 10ms delay to your incoming requests.

RandomInjector

Use fault.RandomInjector to random choose one of the above faults to inject. Pass a list of Injector
to fault.NewRandomInjector and when RandomInjector is evaluated it will randomly insert on of the
injectors that you passed.

Combining Faults

It is easy to combine any of the Injectors into a chained action. There are two ways you might want
to combine Injectors.

First, you can create separate Faults for each Injector that are sequential but independent of each
other. For example, you can chain Faults such that 1% of requests will return a 500 error and
another 1% of requests will be rejected.

Second, you might want to combine Faults such that 1% of requests will be slowed for 10ms and then
rejected. You want these Faults to depend on each other. For this use the special ChainInjector,
which consolidates any number of Injectors into a single Injector that runs each of the provided
Injectors sequentially. When you add the ChainInjector to a Fault the entire chain will always
execute together.

Context Values

The Fault and Injector types are set up to add informational context strings to the request context,
contained in a list with the key fault.ContextKey. The value of that key is a fault.ContextValue
([]ContextString). Once a fault has run you can inspect r.Context().Value(ContextKey) in middleware
further down the chain to observe how the fault was evaluated. For example, ContextValue will
contain ContextString like ContextValueInjected or ContextValueSkipped.

Blacklisting & Whitelisting Paths

The fault.Options struct has PathBlacklist and PathWhitelist options. Any path you include in the
PathBlacklist will never have faults run against in. For PathWhitelist, if you provide a non-empty
list then faults will not be run against any paths except those specified in PathWhitelist. The
PathBlacklist take priority over the PathWhitelist, a path in both lists will never have a fault run
against it. The paths that you include must match exactly the path in req.URL.Path, including
leading and trailing slashes.

Specifying very large lists of paths may cause memory or performance issues. If you're running into
these problems you should instead consider using your http router to enable the middleware on only a
subset of your routes.

Custom Injectors

The package provides an Injector interface and you can satisfy that interface to provide your own
Injector. Use custom injectors to add additional logic (logging, stats) to the package-provided
injectors or to create your own completely new Injector that can still be managed by the Fault
struct.

Configuration

All configuration for the fault package is done through the Options struct. There is no other way to
manage configuration for the package. It is up to the user of the fault package to manage how the
Options struct is generated. Common options are feature flags, environment variables, or code
changes in deploys.

*/
package fault
