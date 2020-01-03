package fault

import (
	"math/rand"
	"net/http"
	"time"
)

const (
	// TypeDrop throws away the request with no response
	TypeDrop = "DROP"
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

	// DROP, REJECT, ERROR, SLOW
	// Use the provided type constants (eg fault.TypeError) to prevent typos
	Type string

	// DROP:  ms before returning an empty reply (default 5 minutes)
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
	// opt holds our configuration for this middleware
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
	if f.opt.Enabled {
		return f.process(h)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// process is the main handler that decides which fault-specific handler
// to call or does nothing if our type is invalid
func (f *Fault) process(h http.Handler) http.Handler {
	switch f.opt.Type {
	case TypeDrop:
		return f.processDrop(h)
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

// processDrop is the handler used when a DROP fault type is provided
func (f *Fault) processDrop(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if percentDo(f.opt.PercentOfRequests) {
			// I don't see a way to truly drop (never respond) an http request within the standard go
			// http stack. An option is http.Hijack() but it does not work on HTTP/2 connections. Dropped
			// requests in the context of fault injection are useful for testing timeouts, and we can
			// simulate that by sleeping for a long (5 minutes) period of time before eventuall returning
			// an empty response.
			if f.opt.Value == 0 {
				time.Sleep(5 * time.Minute)
			} else {
				time.Sleep(time.Duration(f.opt.Value) * time.Millisecond)
			}

			// This is a specialized and documented way of sending an interrupted
			// response to the client without printing the panic stack trace or erroring.
			// https://golang.org/pkg/net/http/#Handler
			panic(http.ErrAbortHandler)
		}

		h.ServeHTTP(w, r)
	})
}

// processReject is the handler used when a REJECT fault type is provided
func (f *Fault) processReject(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if percentDo(f.opt.PercentOfRequests) {
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
		if percentDo(f.opt.PercentOfRequests) {
			// Continue normally if we don't have a valid status code
			if http.StatusText(f.opt.Value) == "" {
				h.ServeHTTP(w, r)
			}

			w.WriteHeader(f.opt.Value)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// processSlow is the handler used when a SLOW fault type is provided
func (f *Fault) processSlow(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if percentDo(f.opt.PercentOfRequests) {
			time.Sleep(time.Duration(f.opt.Value) * time.Millisecond)
		}

		h.ServeHTTP(w, r)
	})
}
