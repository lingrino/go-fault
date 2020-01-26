package fault

import (
	"context"
	"net/http"
)

// updateContextValue takes a request and updates ContextValue (from ContextKey)
// with the provided ContextString and then returns the request again
func updateRequestContextValue(r *http.Request, v ContextString) *http.Request {
	ctx := r.Context()

	val, ok := ctx.Value(ContextKey).(ContextValue)
	if !ok {
		return r.WithContext(context.WithValue(ctx, ContextKey, ContextValue{v}))
	}
	val = append(val, v)
	return r.WithContext(context.WithValue(ctx, ContextKey, val))
}
