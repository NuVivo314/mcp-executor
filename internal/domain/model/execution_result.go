package model

// ExecutionResult holds the outcome of a sandboxed JS execution.
type ExecutionResult struct {
	Output     string // the value returned / printed by the JS code
	StatusCode int    // HTTP status code of the last HTTP call (0 if none)
}
