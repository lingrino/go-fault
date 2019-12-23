package fault

import (
	"fmt"
	"net/http"
)

// Options is a struct for specifying all configuration
// to the fault.Fault middleware
type Options struct {
	// Set to true
	Enabled bool

	// Slow, Error, Drop, Reject
	Type string

	// Slow: ms to wait
	// Error: http error to return
	// Drop: n/a
	// Reject n/a
	Value int

	// How many requests to do this to
	PercentOfRequests float32
}

// Fault is middleware that does fault injection on incoming
// requests based on the configured options
type Fault struct {
	// Add options to determine what happens
	opt Options
}

// New returns a new Fault middleware with the supplied options
func New(o Options) *Fault {
	return &Fault{
		opt: o,
	}
}

// Handler implements http.HandlerFunc to use with net/http
func (f *Fault) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.opt.Enabled {
			stop, err := f.process(w, r)
			if err != nil {
				// Do not exit. Errors should not block
				// TODO - error types, handler here
				fmt.Println(err)
			}
			if stop {
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}

func (f *Fault) process(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	switch f.opt.Type {
	case "ERROR":
		stop, err = f.processError(w, r)
	default:
		return stop, fmt.Errorf("failure type not specified")
	}

	return stop, err
}

func (f *Fault) processError(w http.ResponseWriter, r *http.Request) (bool, error) {
	var stop bool
	var err error

	if percentDo(f.opt.PercentOfRequests) {
		w.WriteHeader(f.opt.Value)
		return true, nil
	}

	return stop, err
}
