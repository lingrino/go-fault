package fault

import (
	"math/rand"
	"net/http"
	"time"
)

const (
	// TypeReject sends an empty reply back
	TypeReject = "REJECT"
	// TypeError injects http errors into the request
	TypeError = "ERROR"
	// TypeSlow injects latency into the request
	TypeSlow = "SLOW"
)

// percentDo takes a percent (0.0 <= per <= 1.0)
// and randomly returns true that percent of the time
func percentDo(p float64) bool {
	var proceed bool

	// bias false if p < 0.0, p > 1.0
	if p > 1.0 || p < 0.0 {
		return false
	}

	// 0.0 <= r < 1.0
	r := rand.Float64()
	if r < p {
		return true
	}

	return proceed
}

// Options is a struct for specifying all configuration
// to the fault.Fault middleware
type Options struct {
	// Set to true to enable the Fault middleware
	Enabled bool

	// REJECT, ERROR, SLOW
	// Use the provided type constants (eg fault.TypeError) to prevent typos
	Type string

	// REJECT n/a
	// ERROR: http error to return
	// SLOW:  ms to wait
	Value int

	// The percent of requests that should have the fault injected.
	// 0.0 <= percent <= 1.0
	PercentOfRequests float64
}

// Fault is middleware that does fault injection on incoming
// requests based on the configured options
type Fault struct {
	// Opt holds our configuration for this middleware
	Opt Options
}

// New returns a new Fault middleware with the supplied options
func New(o Options) *Fault {
	return &Fault{
		Opt: o,
	}
}

// Handler implements http.HandlerFunc to use with net/http
func (f *Fault) Handler(h http.Handler) http.Handler {
	if f.Opt.Enabled {
		return f.process(h)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// process is the main handler that decides which fault-specific handler
// to call or does nothing if our type is invalid
func (f *Fault) process(h http.Handler) http.Handler {
	switch f.Opt.Type {
	case TypeReject:
		return f.processReject(h)
	case TypeError:
		return f.processError(h)
	case TypeSlow:
		return f.processSlow(h)
	default:
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}
}

// processReject is the handler used when a REJECT fault type is provided
func (f *Fault) processReject(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if percentDo(f.Opt.PercentOfRequests) {
			// This is a specialized and documented way of sending an interrupted
			// response to the client without printing the panic stack trace or erroring.
			// https://golang.org/pkg/net/http/#Handler
			panic(http.ErrAbortHandler)
		}

		h.ServeHTTP(w, r)
	})
}

// processError is the handler used when an ERROR fault type is provided
func (f *Fault) processError(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if percentDo(f.Opt.PercentOfRequests) {
			// Continue normally if we don't have a valid status code
			if http.StatusText(f.Opt.Value) == "" {
				h.ServeHTTP(w, r)
			}

			w.WriteHeader(f.Opt.Value)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// processSlow is the handler used when a SLOW fault type is provided
func (f *Fault) processSlow(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if percentDo(f.Opt.PercentOfRequests) {
			time.Sleep(time.Duration(f.Opt.Value) * time.Millisecond)
		}

		h.ServeHTTP(w, r)
	})
}
