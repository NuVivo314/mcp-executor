package model

import "testing"

func TestAPISpec_Fields(t *testing.T) {
	spec := APISpec{
		Name:        "petstore",
		Description: "Pet API",
		BaseURL:     "https://petstore.example.com",
		SpecPath:    "./specs/petstore.yaml",
	}

	if spec.Name != "petstore" {
		t.Errorf("Name = %q, want %q", spec.Name, "petstore")
	}
	if spec.Description != "Pet API" {
		t.Errorf("Description = %q, want %q", spec.Description, "Pet API")
	}
	if spec.BaseURL != "https://petstore.example.com" {
		t.Errorf("BaseURL = %q, want %q", spec.BaseURL, "https://petstore.example.com")
	}
	if spec.SpecPath != "./specs/petstore.yaml" {
		t.Errorf("SpecPath = %q, want %q", spec.SpecPath, "./specs/petstore.yaml")
	}
}

func TestEndpoint_Fields(t *testing.T) {
	ep := Endpoint{
		Path:        "/pets/{petId}",
		Method:      "GET",
		OperationID: "getPetById",
		Summary:     "Get a pet by ID",
		Description: "Returns a single pet",
		Parameters: []Parameter{
			{Name: "petId", In: "path", Required: true, Schema: `{"type":"integer"}`},
		},
		RequestBody: nil,
	}

	if ep.Path != "/pets/{petId}" {
		t.Errorf("Path = %q, want %q", ep.Path, "/pets/{petId}")
	}
	if ep.Method != "GET" {
		t.Errorf("Method = %q, want %q", ep.Method, "GET")
	}
	if ep.OperationID != "getPetById" {
		t.Errorf("OperationID = %q, want %q", ep.OperationID, "getPetById")
	}
	if len(ep.Parameters) != 1 {
		t.Fatalf("len(Parameters) = %d, want 1", len(ep.Parameters))
	}
	if ep.Parameters[0].Name != "petId" {
		t.Errorf("param name = %q, want %q", ep.Parameters[0].Name, "petId")
	}
	if !ep.Parameters[0].Required {
		t.Error("param required = false, want true")
	}
	if ep.RequestBody != nil {
		t.Error("RequestBody should be nil for GET")
	}
}

func TestEndpoint_WithRequestBody(t *testing.T) {
	ep := Endpoint{
		Path:        "/pets",
		Method:      "POST",
		OperationID: "createPet",
		Summary:     "Create a pet",
		RequestBody: &RequestBody{
			Required:    true,
			ContentType: "application/json",
			Schema:      `{"type":"object","properties":{"name":{"type":"string"}}}`,
		},
	}

	if ep.RequestBody == nil {
		t.Fatal("RequestBody should not be nil")
	}
	if !ep.RequestBody.Required {
		t.Error("RequestBody.Required = false, want true")
	}
	if ep.RequestBody.ContentType != "application/json" {
		t.Errorf("ContentType = %q, want %q", ep.RequestBody.ContentType, "application/json")
	}
}

func TestParameter_Locations(t *testing.T) {
	locations := []string{"path", "query", "header", "cookie"}
	for _, loc := range locations {
		p := Parameter{Name: "test", In: loc}
		if p.In != loc {
			t.Errorf("In = %q, want %q", p.In, loc)
		}
	}
}

func TestSearchResult_Fields(t *testing.T) {
	sr := SearchResult{
		Endpoint: Endpoint{
			Path:        "/pets",
			Method:      "GET",
			OperationID: "listPets",
			Summary:     "List all pets",
		},
		Score: 42,
	}

	if sr.Score != 42 {
		t.Errorf("Score = %d, want 42", sr.Score)
	}
	if sr.Endpoint.OperationID != "listPets" {
		t.Errorf("Endpoint.OperationID = %q, want %q", sr.Endpoint.OperationID, "listPets")
	}
}

func TestSearchResult_ZeroScore(t *testing.T) {
	sr := SearchResult{
		Endpoint: Endpoint{Path: "/health"},
		Score:    0,
	}
	if sr.Score != 0 {
		t.Errorf("Score = %d, want 0", sr.Score)
	}
}

func TestExecutionResult_Fields(t *testing.T) {
	er := ExecutionResult{
		Output:     `{"id":1,"name":"Fido"}`,
		StatusCode: 200,
	}

	if er.Output != `{"id":1,"name":"Fido"}` {
		t.Errorf("Output = %q, want JSON", er.Output)
	}
	if er.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", er.StatusCode)
	}
}

func TestExecutionResult_NoHTTPCall(t *testing.T) {
	er := ExecutionResult{
		Output:     "search results here",
		StatusCode: 0,
	}
	if er.StatusCode != 0 {
		t.Errorf("StatusCode = %d, want 0 (no HTTP call)", er.StatusCode)
	}
}
