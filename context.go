package fault

import (
	"context"
	"net/http"
)

// ContextString is the type that all of our context keys will be
type ContextString string

// ContextValue is the value defined by ContextKey. It is a list of ContextString that will be added
// by Injectors
type ContextValue []ContextString

const (
	// ContextKey will be added to the request context of any injector that evaluates and the
	// value will be ContextValue, a list of ContextString that describe what fault occurred.
	ContextKey ContextString = "fault"
	// ContextValueError is added to ContextValue when an error (ex: misconfiguration) occurred
	// while trying to inject a fault
	ContextValueError ContextString = "fault-error"

	// ContextValueInjected is added at the Fault level to any request that is evaluated
	ContextValueInjected = "fault-injected"
	// ContextValueDisabled is added to ContextValue when the fault is disabled
	ContextValueDisabled ContextString = "fault-disabled"
	// ContextValueDisabled is added to ContextValue when the fault is skipped (not in
	// PercentOfRequests)
	ContextValueSkipped ContextString = "fault-skipped"

	// ContextValueChainInjector is added to ContextValue when the ChainInjector is injected
	ContextValueChainInjector ContextString = "chain-injector"
	// ContextValueRandomInjector is added to ContextValue when the RandomInjector is injected
	ContextValueRandomInjector ContextString = "random-injector"
	// ContextValueSlowInjector is added to ContextValue when the SlowInjector is injected
	ContextValueSlowInjector ContextString = "slow-injector"
)

// updateRequestContextValue takes a request and updates ContextValue (from ContextKey) with the provided
// ContextString and then returns a shallow copy of the request
func updateRequestContextValue(r *http.Request, v ContextString) *http.Request {
	if r != nil {
		ctx := r.Context()

		val, ok := ctx.Value(ContextKey).(ContextValue)
		if !ok {
			return r.WithContext(context.WithValue(ctx, ContextKey, ContextValue{v}))
		}
		val = append(val, v)
		return r.WithContext(context.WithValue(ctx, ContextKey, val))
	}
	return r
}
