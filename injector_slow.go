package fault

import (
	"net/http"
	"time"
)

// SlowInjector sleeps a specified duration and then continues the request. Simulates latency.
type SlowInjector struct {
	duration time.Duration
	sleep    func(t time.Duration)
}

// NewSlowInjector returns a SlowInjector that adds the configured latency.
func NewSlowInjector(d time.Duration) (*SlowInjector, error) {
	return &SlowInjector{
		duration: d,
		sleep:    time.Sleep,
	}, nil
}

// Handler waits the configured duration and then continues the request.
func (i *SlowInjector) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i != nil {
			if i.sleep != nil {
				i.sleep(i.duration)
			}
		}
		next.ServeHTTP(w, r)
	})
}
