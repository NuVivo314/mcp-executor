package port

import "context"

// Sandbox executes JavaScript code in an isolated runtime.
type Sandbox interface {
	// EvalSearch runs JS code with search helpers injected (search, getEndpoints, getSpec).
	EvalSearch(ctx context.Context, code, apiName string) (string, error)

	// EvalExecute runs JS code with HTTP helpers injected (httpGet, httpPost, …) and search helpers.
	EvalExecute(ctx context.Context, code, apiName string) (string, error)
}
