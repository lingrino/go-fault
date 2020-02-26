package fault

import (
	"log"
)

// Reporter receives event data from injected faults to use for logging, stats, and other custom
// reporting.
type Reporter interface {
	Report(name string, state InjectorState)
}

// StandardReporter is a simple reporter that logs all reports. It does very little because reporters
// are meant to be implemented by clients for custom observability requirements like logging or
// stats reporting.
type StandardReporter struct{}

// NewStandard returns a new Standard
func NewStandardReporter() *StandardReporter {
	return &StandardReporter{}
}

// Report simply logs the input message using the standard go logger
func (r *StandardReporter) Report(name string, state InjectorState) {
	if r != nil {
		log.Println(name, "-", state)
	}
}

// NoopReporter is a reporter that does nothing.
type NoopReporter struct{}

// NewNoopReporter returns a new NoopReporter
func NewNoopReporter() *NoopReporter {
	return &NoopReporter{}
}

// Report does nothing
func (r *NoopReporter) Report(name string, state InjectorState) {}

// reportWithMessage is a helper function to simplify sending simple messages
func reportWithMessage(r Reporter, name string, state InjectorState) {
	if r != nil {
		r.Report(name, state)
	}
}
