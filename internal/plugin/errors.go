package plugin

import "fmt"

// ErrABIVersion is returned when a plugin declares an unsupported ABI version.
type ErrABIVersion struct {
	Got  int32
	Want int32
}

func (e *ErrABIVersion) Error() string {
	return fmt.Sprintf("plugin: unsupported ABI version %d (host supports %d)", e.Got, e.Want)
}

// ErrMissingExport is returned when a required export function is absent.
type ErrMissingExport struct {
	Name string
}

func (e *ErrMissingExport) Error() string {
	return fmt.Sprintf("plugin: missing required export %q", e.Name)
}

// ErrExecution wraps a runtime error that occurred during hook execution.
type ErrExecution struct {
	Cause error
}

func (e *ErrExecution) Error() string {
	return fmt.Sprintf("plugin: execution error: %v", e.Cause)
}

func (e *ErrExecution) Unwrap() error { return e.Cause }

// ErrTimeout is returned when a plugin hook exceeds its time limit.
type ErrTimeout struct{}

func (e *ErrTimeout) Error() string { return "plugin: hook execution timed out" }

// ErrInvalidHandle is returned when a plugin uses a handle that is not valid.
type ErrInvalidHandle struct {
	Handle int32
}

func (e *ErrInvalidHandle) Error() string {
	return fmt.Sprintf("plugin: invalid or expired handle %d", e.Handle)
}
