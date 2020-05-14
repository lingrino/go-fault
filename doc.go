/*
Package fault provides standard http middleware for fault injection in go.

Basics

Use the fault package to inject faults into the http request path of your service. Faults work by
modifying and/or delaying your service's http responses. Place the Fault middleware high enough in
the chain that it can act quickly, but after any other middlewares that should complete before fault
injection (auth, redirects, etc...).

The type and severity of injected faults is controlled by options passed to NewFault(Injector,
Options). NewFault must be passed an Injector, which is an interface that holds the actual fault
injection code in Injector.Handler. The Fault wraps Injector.Handler in another Fault.Handler that
applies generic Fault logic (such as what % of requests to run the Injector on) to the Injector.

Make sure you use the NewFault() and NewTypeInjector() constructors to create valid Faults and
Injectors.

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

Use fault.ErrorInjector to immediately return a valid http status code of your choosing along with
the standard HTTP response body for that code. For example, you can return a 200, 301, 418, 500, or
any other valid status code to test how your clients respond to different statuses. Pass the
WithStatusText() option to customize the response text.

SlowInjector

Use fault.SlowInjector to wait a configured time.Duration before proceeding with the request. For
example, you can use the SlowInjector to add a 10ms delay to your requests.

RandomInjector

Use fault.RandomInjector to randomly choose one of the above faults to inject. Pass a list of
Injector to fault.NewRandomInjector and when RandomInjector is evaluated it will randomly run one of
the injectors that you passed.

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

Allowing & Blocking Paths

The NewFault() constructor has WithPathBlocklist() and WithPathAllowlist() options. Any path you
include in the PathBlocklist will never have faults run against it. With PathAllowlist, if you
provide a non-empty list then faults will not be run against any paths except those specified in
PathAllowlist. The PathBlocklist take priority over the PathAllowlist, a path in both lists will
never have a fault run against it. The paths that you include must match exactly the path in
req.URL.Path, including leading and trailing slashes.

Simmilarly, you may also use WithHeaderBlocklist() and WithHeaderAllowlist() to block or allow
faults based on a map of header keys to values. These lists behave in the same way as the path
allowlists and blocklists except that they operate on headers. Header equality is determined using
http.Header.Get(key) which automatically canonicalizes your keys and does not support multi-value
headers. Keep these limitations in mind when working with header allowlists and blocklists.

Specifying very large lists of paths or headers may cause memory or performance issues. If you're
running into these problems you should instead consider using your http router to enable the
middleware on only a subset of your routes.

Custom Injectors

The fault package provides an Injector interface and you can satisfy that interface to provide your
own Injector. Use custom injectors to add additional logic to the package-provided injectors or to
create your own completely new Injector that can still be managed by a Fault.

Reporter

The package provides a Reporter interface that can be added to Faults and Injectors using the
WithReporter option. A Reporter will receive events when the state of the Injector changes. For
example, Reporter.Report(InjectorName, StateStarted) is run at the beginning of all Injectors. The
Reporter is meant to be provided by the consumer of the package and integrate with services like
stats and logging. The default Reporter throws away all events.

Random Seeds

By default all randomness is seeded with defaultRandSeed(1), the same default as math/rand. This
helps you reproduce any errors you see when running an Injector. If you prefer, you can also
customize the seed passing WithRandSeed() to NewFault and NewRandomInjector.

Custom Injector Functions

Some Injectors support customizing the functions they use to run their injections. You can take
advantage of these options to add your own logic to an existing Injector instead of creating your
own. For example, modify the SlowInjector function to slow in a rancom distribution instead of for a
fixed duration. Be careful when you use these options that your return values fall within the same
range of values expected by the default functions to avoid panics or other undesirable begavior.

Customize the function a Fault uses to determine participation (default: rand.Float32) by passing
WithRandFloat32Func() to NewFault().

Customize the function a RandomInjector uses to choose which injector to run (default: rand.Intn) by
passing WithRandIntFunc() to NewRandomInjector().

Customize the function a SlowInjector uses to wait (default: time.Sleep) by passing WithSlowFunc()
to NewSlowInjector().

Configuration

All configuration for the fault package is done through options passed to NewFault and NewInjector.
There is no other way to manage configuration for the package. It is up to the user of the fault
package to manage how the options are generated. Common options are feature flags, environment
variables, or code changes in deploys.

*/
package fault
