package fault

import (
	"net/http"
)

// InjectorState represents the states an injector can be in.
type InjectorState int

const (
	StateStarted InjectorState = iota + 1
	StateFinished
	StateSkipped
)

// Injector is an interface for our fault injection middleware. Injectors are wrapped into Faults.
// Faults handle running the Injector the correct percent of the time.
type Injector interface {
	Handler(next http.Handler) http.Handler
}
