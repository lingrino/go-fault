package fault

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Type is the type of fault injection (REJECT, ERROR, SLOW) to be run
type Type string

const (
	// Reject sends an empty reply back
	Reject Type = "REJECT"
	// Error injects http errors into the request
	Error Type = "ERROR"
	// Slow injects latency into the request
	Slow Type = "SLOW"

	// FaultInjected is passed in the request context of all non-returning (SLOW)
	// fault types. This value supports the Chained functionality.
	FaultInjected Type = ""
)

// Options is a struct for specifying all configuration
// to the fault.Fault middleware
type Options struct {
	// Enabled determines if the fault middleware should evaluate
	Enabled bool

	// Type is the type of fault to be injected
	// Options - Reject, Error, Slow
	Type Type

	// Value represents how the fault should act. The meaning changes
	// for different kinds of faults.
	// REJECT n/a
	// ERROR: http error to return
	// SLOW:  ms to wait
	Value uint

	// PercentOfRequests is the percent of requests that should have the fault injected.
	// 0.0 <= percent <= 1.0
	PercentOfRequests float32

	// Chained determines if this fault should only activate when a non-returning (SLOW)
	// fault higher up the chain has activated. This ignores PercentOfRequests.
	// Use to chain faults like 20% SLOW then REJECT.
	Chained bool

	// Debug determines if we should log errors and fault occurences. Requests will always
	// continue without injecting a fault if the fault is misconfigured.
	Debug bool

	// Logger is your custom logger. Defaults to the standard log.Logger
	Logger *log.Logger
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

// log is a wrapper for logging our debug messages only when debug is enabled
// and choosing the provided or standard logger.
func (f *Fault) log(v ...interface{}) {
	if f.Opt.Debug {
		if f.Opt.Logger != nil {
			f.Opt.Logger.Println(v...)
		} else {
			log.Println(v...)
		}
	}
}

// percentDo takes a percent (0.0 <= per <= 1.0)
// and randomly returns true that percent of the time
func (f *Fault) percentDo(r *http.Request) bool {
	var proceed bool

	// Always proceed on chained requests where an earlier fault
	// has already been injected. Do nothing if fault.Chained
	// but no earlier fault was injected.
	if f.Opt.Chained {
		if r.Context().Value(FaultInjected) != nil {
			return true
		}
		return false
	}

	// bias false if p < 0.0, p > 1.0
	if f.Opt.PercentOfRequests > 1.0 || f.Opt.PercentOfRequests < 0.0 {
		f.log("PercentOfRequests must be 0.0 <= per <= 1.0, got:", f.Opt.PercentOfRequests)
		return false
	}

	// 0.0 <= r < 1.0
	rn := rand.Float32()
	if rn < f.Opt.PercentOfRequests {
		return true
	}

	return proceed
}

// process is the main handler that decides which fault-specific handler
// to call or does nothing if our type is invalid
func (f *Fault) process(h http.Handler) http.Handler {
	switch f.Opt.Type {
	case Reject:
		return f.processReject(h)
	case Error:
		return f.processError(h)
	case Slow:
		return f.processSlow(h)
	default:
		f.log("fault type is not valid. got:", f.Opt.Type)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	}
}

// processReject is the handler used when a REJECT fault type is provided
func (f *Fault) processReject(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.percentDo(r) {
			// This is a specialized and documented way of sending an interrupted
			// response to the client without printing the panic stack trace or erroring.
			// https://golang.org/pkg/net/http/#Handler
			panic(http.ErrAbortHandler)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// processError is the handler used when an ERROR fault type is provided
func (f *Fault) processError(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.percentDo(r) {
			statusText := http.StatusText(int(f.Opt.Value))
			// Continue normally if we don't have a valid status code
			if statusText == "" {
				f.log(f.Opt.Value, "is not a valid http status code")
				h.ServeHTTP(w, r)
			} else {
				http.Error(w, statusText, int(f.Opt.Value))
			}

			return
		}

		h.ServeHTTP(w, r)
	})
}

// processSlow is the handler used when a SLOW fault type is provided
func (f *Fault) processSlow(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.percentDo(r) {
			time.Sleep(time.Duration(f.Opt.Value) * time.Millisecond)
			ctx := context.WithValue(r.Context(), FaultInjected, Slow)

			h.ServeHTTP(w, r.WithContext(ctx))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
