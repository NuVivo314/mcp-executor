package port

import "github.com/nuvivo314/mcp-executor/internal/domain/model"

// SearchEngine performs fuzzy search over API endpoints.
type SearchEngine interface {
	// Search returns endpoints matching query within the named API.
	Search(apiName, query string) ([]model.SearchResult, error)
}
