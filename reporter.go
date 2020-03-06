package fault

// Reporter receives events from faults to use for logging, stats, and other custom reporting.
type Reporter interface {
	Report(name string, state InjectorState)
}

// NoopReporter is a reporter that does nothing.
type NoopReporter struct{}

// NewNoopReporter returns a new NoopReporter.
func NewNoopReporter() *NoopReporter {
	return &NoopReporter{}
}

// Report does nothing.
func (r *NoopReporter) Report(name string, state InjectorState) {}

// ReporterOption configures structs that accept a Reporter.
type ReporterOption interface {
	RejectInjectorOption
	ErrorInjectorOption
	SlowInjectorOption
}

// reporterOption holds our passed in Reporter.
type reporterOption struct {
	reporter Reporter
}

// WithReporter sets the Reporter.
func WithReporter(r Reporter) ReporterOption {
	return reporterOption{r}
}
