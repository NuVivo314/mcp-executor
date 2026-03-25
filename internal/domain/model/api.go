package model

// APISpec describes a configured API.
type APISpec struct {
	Name        string
	Description string
	BaseURL     string
	SpecPath    string
}

// Endpoint represents a single HTTP endpoint from an OpenAPI spec.
type Endpoint struct {
	Path        string
	Method      string
	OperationID string
	Summary     string
	Description string
	Parameters  []Parameter
	RequestBody *RequestBody
}

// Parameter represents a single OpenAPI parameter (path, query, header, cookie).
type Parameter struct {
	Name     string
	In       string // "path" | "query" | "header" | "cookie"
	Required bool
	Schema   string // JSON-serialised schema snippet
}

// RequestBody describes the request body schema.
type RequestBody struct {
	Required    bool
	ContentType string
	Schema      string // JSON-serialised schema snippet
}
