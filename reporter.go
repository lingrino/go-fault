package fault

import (
	"log"
	"net/http"
)

// Reporter receives event data from injected faults to use for logging, stats, and other custom reporting.
type Reporter interface {
	Report(*http.Request) error
}

type StandardReporter struct{}

func (s *StandardReporter) Report(r *http.Request) error {
	ctx := r.Context().Value(ContextKey)
	log.Println(ctx)
	return nil
}
