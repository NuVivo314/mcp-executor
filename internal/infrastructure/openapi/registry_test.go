package openapi

import (
	"testing"

	"github.com/nuvivo314/mcp-executor/internal/config"
)

func testConfigs() []config.APIConfig {
	return []config.APIConfig{
		{
			Name:        "petstore",
			Description: "Petstore sample API",
			SpecPath:    "testdata/petstore.yaml",
			BaseURL:     "https://petstore.example.com",
		},
	}
}

func TestNewRegistry_Valid(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg == nil {
		t.Fatal("registry should not be nil")
	}
}

func TestNewRegistry_MissingSpec(t *testing.T) {
	cfgs := []config.APIConfig{
		{Name: "bad", SpecPath: "testdata/nonexistent.yaml", BaseURL: "https://example.com"},
	}
	_, err := NewRegistry(cfgs)
	if err == nil {
		t.Fatal("expected error for missing spec file")
	}
}

func TestListAPIs(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	apis := reg.ListAPIs()
	if len(apis) != 1 {
		t.Fatalf("len(apis) = %d, want 1", len(apis))
	}
	if apis[0].Name != "petstore" {
		t.Errorf("name = %q, want %q", apis[0].Name, "petstore")
	}
	if apis[0].BaseURL != "https://petstore.example.com" {
		t.Errorf("base_url = %q, want %q", apis[0].BaseURL, "https://petstore.example.com")
	}
}

func TestGetAPI_Found(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	spec, err := reg.GetAPI("petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Name != "petstore" {
		t.Errorf("name = %q, want %q", spec.Name, "petstore")
	}
}

func TestGetAPI_NotFound(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err = reg.GetAPI("unknown")
	if err == nil {
		t.Fatal("expected error for unknown API")
	}
}

func TestGetEndpoints_Count(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	endpoints, err := reg.GetEndpoints("petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// petstore has 3 operations: GET /pets, POST /pets, GET /pets/{petId}
	if len(endpoints) != 3 {
		t.Fatalf("len(endpoints) = %d, want 3", len(endpoints))
	}
}

func TestGetEndpoints_Fields(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	endpoints, err := reg.GetEndpoints("petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find GET /pets/{petId}
	var ep *struct {
		path   string
		method string
		opID   string
	}
	for _, e := range endpoints {
		if e.OperationID == "getPetById" {
			ep = &struct {
				path   string
				method string
				opID   string
			}{e.Path, e.Method, e.OperationID}
			break
		}
	}
	if ep == nil {
		t.Fatal("endpoint getPetById not found")
	}
	if ep.path != "/pets/{petId}" {
		t.Errorf("path = %q, want %q", ep.path, "/pets/{petId}")
	}
	if ep.method != "GET" {
		t.Errorf("method = %q, want %q", ep.method, "GET")
	}
}

func TestGetEndpoints_Parameters(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	endpoints, err := reg.GetEndpoints("petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range endpoints {
		if e.OperationID == "getPetById" {
			if len(e.Parameters) != 1 {
				t.Fatalf("getPetById params = %d, want 1", len(e.Parameters))
			}
			p := e.Parameters[0]
			if p.Name != "petId" {
				t.Errorf("param name = %q, want %q", p.Name, "petId")
			}
			if p.In != "path" {
				t.Errorf("param in = %q, want %q", p.In, "path")
			}
			if !p.Required {
				t.Error("petId should be required")
			}
			return
		}
	}
	t.Fatal("getPetById not found")
}

func TestGetEndpoints_RequestBody(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	endpoints, err := reg.GetEndpoints("petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range endpoints {
		if e.OperationID == "createPet" {
			if e.RequestBody == nil {
				t.Fatal("createPet should have a request body")
			}
			if !e.RequestBody.Required {
				t.Error("request body should be required")
			}
			if e.RequestBody.ContentType != "application/json" {
				t.Errorf("content type = %q, want application/json", e.RequestBody.ContentType)
			}
			return
		}
	}
	t.Fatal("createPet not found")
}

func TestGetEndpoints_UnknownAPI(t *testing.T) {
	reg, err := NewRegistry(testConfigs())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err = reg.GetEndpoints("unknown")
	if err == nil {
		t.Fatal("expected error for unknown API")
	}
}
