package fault

// Reporter receives event data from injected faults to use for logging, stats, and other custom
// reporting.
type Reporter interface {
	Report(name string, state InjectorState)
}

// NoopReporter is a reporter that does nothing.
type NoopReporter struct{}

// NewNoopReporter returns a new NoopReporter
func NewNoopReporter() *NoopReporter {
	return &NoopReporter{}
}

// Report does nothing
func (r *NoopReporter) Report(name string, state InjectorState) {}
