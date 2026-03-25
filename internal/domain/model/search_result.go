package model

// SearchResult represents a single fuzzy-search match against API endpoints.
type SearchResult struct {
	Endpoint Endpoint
	Score    int
}
