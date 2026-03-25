package port

import "github.com/nuvivo314/mcp-executor/internal/domain/model"

// ApiRegistry provides access to configured APIs and their endpoints.
type ApiRegistry interface {
	// ListAPIs returns all registered API specs.
	ListAPIs() []model.APISpec

	// GetAPI returns the spec for the named API, or an error if not found.
	GetAPI(name string) (*model.APISpec, error)

	// GetEndpoints returns all endpoints for the named API.
	GetEndpoints(apiName string) ([]model.Endpoint, error)
}
