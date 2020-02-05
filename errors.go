package fault

import "errors"

var (
	// ErrNilInjector returns when a nil Injector type is passed.
	ErrNilInjector = errors.New("injector cannot be nil")
	// ErrInvalidPercent returns when a provided percent is outside of the allowed bounds.
	ErrInvalidPercent = errors.New("percent must be 0.0 <= percent <= 1.0")
	// ErrInvalidHTTPCode returns when an invalid http status code is provided.
	ErrInvalidHTTPCode = errors.New("not a valid http status code")
)
