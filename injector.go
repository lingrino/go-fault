package fault

import (
	"net/http"
)

// InjectorState represents the states an injector can be in.
type InjectorState int

const (
	// StateStarted when an Injector has started.
	StateStarted InjectorState = iota + 1
	// StateFinished when an Injector has finished.
	StateFinished
)

// Injector are added to Faults and run as middleware in a request.
type Injector interface {
	Handler(next http.Handler) http.Handler
}
