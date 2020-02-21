package fault

import (
	"log"
	"net/http"
)

// Reporter receives event data from injected faults to use for logging, stats, and other custom
// reporting.
type Reporter interface {
	Report(ReportInput)
}

// ReportInput holds all relevant event data for Reporter.Report
type ReportInput struct {
	// A message that describes the event
	Msg string
	// The http request that is being described
	Req *http.Request
}

// DefaultReporter is the default Reporter for the package. It does very little because reporters
// are meant to be implemented by clients for custom observability requirements like logging or
// stats reporting.
type DefaultReporter struct{}

// Report simply logs the input message using the standard go logger
func (r *DefaultReporter) Report(i ReportInput) {
	if r != nil {
		log.Println(i.Msg)
	}
}
