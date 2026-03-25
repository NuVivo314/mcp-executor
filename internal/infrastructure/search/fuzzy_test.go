package search

import (
	"errors"
	"testing"

	"github.com/nuvivo314/mcp-executor/internal/domain/model"
	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

// stubRegistry implements port.ApiRegistry for testing.
type stubRegistry struct {
	endpoints map[string][]model.Endpoint
}

func (s *stubRegistry) ListAPIs() []model.APISpec { return nil }

func (s *stubRegistry) GetAPI(name string) (*model.APISpec, error) {
	if _, ok := s.endpoints[name]; !ok {
		return nil, errors.New("not found")
	}
	return &model.APISpec{Name: name}, nil
}

func (s *stubRegistry) GetEndpoints(apiName string) ([]model.Endpoint, error) {
	eps, ok := s.endpoints[apiName]
	if !ok {
		return nil, errors.New("api not found")
	}
	return eps, nil
}

var _ port.ApiRegistry = (*stubRegistry)(nil)

func petstoreRegistry() *stubRegistry {
	return &stubRegistry{
		endpoints: map[string][]model.Endpoint{
			"petstore": {
				{Path: "/pets", Method: "GET", OperationID: "listPets", Summary: "List all pets"},
				{Path: "/pets", Method: "POST", OperationID: "createPet", Summary: "Create a pet"},
				{Path: "/pets/{petId}", Method: "GET", OperationID: "getPetById", Summary: "Get pet by ID"},
			},
		},
	}
}

func TestNewFuzzySearch(t *testing.T) {
	se := NewFuzzySearch(petstoreRegistry())
	if se == nil {
		t.Fatal("search engine should not be nil")
	}
}

func TestSearch_ReturnsResults(t *testing.T) {
	se := NewFuzzySearch(petstoreRegistry())
	results, err := se.Search("petstore", "pet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for query 'pet'")
	}
}

func TestSearch_RelevantMatch(t *testing.T) {
	se := NewFuzzySearch(petstoreRegistry())
	results, err := se.Search("petstore", "listPets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 'listPets'")
	}
	// The top result should be listPets
	found := false
	for _, r := range results {
		if r.Endpoint.OperationID == "listPets" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("listPets not found in results: %v", results)
	}
}

func TestSearch_UnknownAPI(t *testing.T) {
	se := NewFuzzySearch(petstoreRegistry())
	_, err := se.Search("unknown", "pet")
	if err == nil {
		t.Fatal("expected error for unknown API")
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	se := NewFuzzySearch(petstoreRegistry())
	results, err := se.Search("petstore", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty query should return all endpoints
	if len(results) != 3 {
		t.Errorf("empty query: got %d results, want 3", len(results))
	}
}

func TestSearch_ScoresPresent(t *testing.T) {
	se := NewFuzzySearch(petstoreRegistry())
	results, err := se.Search("petstore", "create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for 'create'")
	}
	for _, r := range results {
		if r.Score < 0 {
			t.Errorf("score should be >= 0, got %d", r.Score)
		}
	}
}
