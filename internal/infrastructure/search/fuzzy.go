package search

import (
	"fmt"

	"github.com/nuvivo314/mcp-executor/internal/domain/model"
	"github.com/nuvivo314/mcp-executor/internal/domain/port"
	"github.com/sahilm/fuzzy"
)

type fuzzySearch struct {
	registry port.ApiRegistry
}

// NewFuzzySearch returns a SearchEngine backed by sahilm/fuzzy.
func NewFuzzySearch(registry port.ApiRegistry) port.SearchEngine {
	return &fuzzySearch{registry: registry}
}

// endpointSource adapts []model.Endpoint to the fuzzy.Source interface.
type endpointSource struct {
	endpoints []model.Endpoint
}

func (s endpointSource) String(i int) string {
	ep := s.endpoints[i]
	// Combine fields so all are searchable
	return ep.OperationID + " " + ep.Summary + " " + ep.Path + " " + ep.Method
}

func (s endpointSource) Len() int { return len(s.endpoints) }

func (f *fuzzySearch) Search(apiName, query string) ([]model.SearchResult, error) {
	endpoints, err := f.registry.GetEndpoints(apiName)
	if err != nil {
		return nil, fmt.Errorf("getting endpoints for %q: %w", apiName, err)
	}

	// Empty query: return all endpoints with score 0.
	if query == "" {
		results := make([]model.SearchResult, len(endpoints))
		for i, ep := range endpoints {
			results[i] = model.SearchResult{Endpoint: ep, Score: 0}
		}
		return results, nil
	}

	src := endpointSource{endpoints: endpoints}
	matches := fuzzy.FindFrom(query, src)

	results := make([]model.SearchResult, len(matches))
	for i, m := range matches {
		results[i] = model.SearchResult{
			Endpoint: endpoints[m.Index],
			Score:    m.Score,
		}
	}
	return results, nil
}
