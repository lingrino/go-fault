# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0] - Unreleased

### Breaking Changes

#### API Changes

- **`ErrInvalidPercent` renamed to `ErrInvalidParticipation`**

  Update any error checks:
  ```go
  // Before
  if err == fault.ErrInvalidPercent { ... }

  // After
  if err == fault.ErrInvalidParticipation { ... }
  ```

- **`SetEnabled` now takes `bool` directly**

  ```go
  // Before
  f.SetEnabled(fault.WithEnabled(true))

  // After
  f.SetEnabled(true)
  ```

- **`SetParticipation` now takes `float32` directly**

  ```go
  // Before
  f.SetParticipation(fault.WithParticipation(0.5))

  // After
  f.SetParticipation(0.5)
  ```

- **`NewChainInjector` no longer accepts options**

  ```go
  // Before
  ci, err := fault.NewChainInjector(injectors, opts...)

  // After
  ci, err := fault.NewChainInjector(injectors)
  ```

- **`NewRandomInjector` now errors on empty/nil slice**

  Previously, passing an empty or nil slice would succeed and create a no-op injector. Now it returns `ErrEmptyInjectorSlice`. Ensure you pass at least one injector.

- **`StateSkipped` constant removed**

  This constant was defined but never used. Remove any references to it.

- **Header allowlist now matches ANY header instead of ALL**

  Previously, `WithHeaderAllowlist` required ALL specified headers to match. Now it matches if ANY header matches, which is consistent with how `WithHeaderBlocklist` works. If you relied on the ALL-match behavior, you'll need to update your logic.

### Bug Fixes

- **RejectInjector now correctly reports `StateFinished`** before panicking with `http.ErrAbortHandler`. Previously, the `StateFinished` event was never reported.

- **Thread-safety added to `SetEnabled()` and `SetParticipation()`**. These methods are now safe to call concurrently with `Handler()`.

### Improvements

- **Nil injector validation**: `NewChainInjector` and `NewRandomInjector` now return `ErrNilInjector` if any injector in the slice is nil.

### Documentation

- Fixed typos in doc.go
- Added CLAUDE.md for AI assistant guidance
