package sandbox

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nuvivo314/mcp-executor/internal/domain/model"
	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

// stubRegistry for sandbox tests.
type stubRegistry struct{}

func (s *stubRegistry) ListAPIs() []model.APISpec { return nil }
func (s *stubRegistry) GetAPI(name string) (*model.APISpec, error) {
	if name == "petstore" {
		return &model.APISpec{Name: "petstore", BaseURL: "https://petstore.example.com"}, nil
	}
	return nil, errors.New("not found")
}
func (s *stubRegistry) GetEndpoints(apiName string) ([]model.Endpoint, error) {
	if apiName != "petstore" {
		return nil, errors.New("not found")
	}
	return []model.Endpoint{
		{Path: "/pets", Method: "GET", OperationID: "listPets", Summary: "List all pets"},
		{Path: "/pets/{petId}", Method: "GET", OperationID: "getPetById", Summary: "Get pet by ID"},
	}, nil
}

// stubSearchEngine for sandbox tests.
type stubSearchEngine struct{}

func (s *stubSearchEngine) Search(apiName, query string) ([]model.SearchResult, error) {
	return []model.SearchResult{
		{Endpoint: model.Endpoint{OperationID: "listPets", Summary: "List all pets"}, Score: 10},
	}, nil
}

// stubHTTPClient for sandbox tests.
type stubHTTPClient struct {
	response *port.HttpResponse
	err      error
}

func (s *stubHTTPClient) Do(_ context.Context, method, url string, headers map[string]string, body string) (*port.HttpResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.response != nil {
		return s.response, nil
	}
	return &port.HttpResponse{StatusCode: 200, Body: `{"ok":true}`}, nil
}

func newTestSandbox() *Sandbox {
	return NewSandbox(
		&stubRegistry{},
		&stubSearchEngine{},
		&stubHTTPClient{},
	)
}

// --- EvalSearch tests ---

func TestEvalSearch_BasicJS(t *testing.T) {
	sb := newTestSandbox()
	result, err := sb.EvalSearch(context.Background(), `1 + 1`, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "2" {
		t.Errorf("result = %q, want %q", result, "2")
	}
}

func TestEvalSearch_SearchHelper(t *testing.T) {
	sb := newTestSandbox()
	code := `JSON.stringify(search("pets"))`
	result, err := sb.EvalSearch(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "listPets") {
		t.Errorf("result = %q, want to contain 'listPets'", result)
	}
}

func TestEvalSearch_GetEndpointsHelper(t *testing.T) {
	sb := newTestSandbox()
	code := `JSON.stringify(getEndpoints())`
	result, err := sb.EvalSearch(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "listPets") {
		t.Errorf("result should contain endpoint data, got: %q", result)
	}
}

func TestEvalSearch_GetSpecHelper(t *testing.T) {
	sb := newTestSandbox()
	code := `typeof getSpec()`
	result, err := sb.EvalSearch(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "string" {
		t.Errorf("getSpec() should return a string, got type: %q", result)
	}
}

func TestEvalSearch_SyntaxError(t *testing.T) {
	sb := newTestSandbox()
	_, err := sb.EvalSearch(context.Background(), `{{{`, "petstore")
	if err == nil {
		t.Fatal("expected error for syntax error")
	}
}

func TestEvalSearch_UnknownAPI(t *testing.T) {
	sb := newTestSandbox()
	_, err := sb.EvalSearch(context.Background(), `1`, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown API")
	}
}

func TestEvalSearch_ContextTimeout(t *testing.T) {
	sb := newTestSandbox()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	// Infinite loop should be stopped by timeout
	code := `while(true){}`
	_, err := sb.EvalSearch(ctx, code, "petstore")
	if err == nil {
		t.Fatal("expected error for context timeout")
	}
}

// --- EvalExecute tests ---

func TestEvalExecute_BasicJS(t *testing.T) {
	sb := newTestSandbox()
	result, err := sb.EvalExecute(context.Background(), `"hello"`, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("result = %q, want %q", result, "hello")
	}
}

func TestEvalExecute_HTTPGetHelper(t *testing.T) {
	sb := newTestSandbox()
	code := `JSON.stringify(httpGet("/pets"))`
	result, err := sb.EvalExecute(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "statusCode") {
		t.Errorf("result = %q, want to contain 'statusCode'", result)
	}
}

func TestEvalExecute_HTTPPostHelper(t *testing.T) {
	sb := newTestSandbox()
	code := `JSON.stringify(httpPost("/pets", JSON.stringify({name:"Rex"})))`
	result, err := sb.EvalExecute(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "statusCode") {
		t.Errorf("result = %q, want to contain 'statusCode'", result)
	}
}

func TestEvalExecute_SearchHelperAvailable(t *testing.T) {
	sb := newTestSandbox()
	code := `typeof search`
	result, err := sb.EvalExecute(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "function" {
		t.Errorf("search should be a function in execute context, got: %q", result)
	}
}

func TestEvalExecute_NoFilesystemAccess(t *testing.T) {
	sb := newTestSandbox()
	// QuickJS has no fs module — attempting to use it should fail or return undefined
	code := `typeof require`
	result, err := sb.EvalExecute(context.Background(), code, "petstore")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "function" {
		t.Error("require should not be available in sandbox")
	}
}
