package port

import "context"

// HttpClient executes HTTP requests on behalf of the sandbox.
// Credentials are injected by the caller; the JS layer never sees them.
type HttpClient interface {
	// Do performs an HTTP request and returns the response body as a string.
	Do(ctx context.Context, method, url string, headers map[string]string, body string) (*HttpResponse, error)
}

// HttpResponse holds the result of an HTTP call.
type HttpResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}
