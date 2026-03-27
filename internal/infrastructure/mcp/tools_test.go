package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/nuvivo314/mcp-executor/internal/domain/model"
)

// --- stubs ---

type stubListApi struct {
	apis []model.APISpec
}

func (s *stubListApi) Execute() []model.APISpec { return s.apis }

type stubSearch struct {
	result string
	err    error
}

func (s *stubSearch) Execute(ctx context.Context, apiName, code string) (string, error) {
	return s.result, s.err
}

type stubExecute struct {
	result string
	err    error
}

func (s *stubExecute) Execute(ctx context.Context, apiName, code string) (string, error) {
	return s.result, s.err
}

// --- helpers ---

func makeRequest(args map[string]any) mcp.CallToolRequest {
	raw, _ := json.Marshal(args)
	var req mcp.CallToolRequest
	_ = json.Unmarshal([]byte(`{"params":{"arguments":`+string(raw)+`}}`), &req)
	return req
}

func extractText(r *mcp.CallToolResult) string {
	if r == nil || len(r.Content) == 0 {
		return ""
	}
	return mcp.GetTextFromContent(r.Content[0])
}

// --- list_api resource ---

func TestHandleListAPIResource_Empty(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{}, &stubExecute{})
	contents, err := h.HandleListAPIResource(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contents) == 0 {
		t.Fatal("expected at least one resource content")
	}
}

func TestHandleListAPIResource_ReturnsList(t *testing.T) {
	apis := []model.APISpec{
		{Name: "petstore", Description: "Pet API", BaseURL: "https://petstore.example.com"},
	}
	h := NewHandlers(&stubListApi{apis: apis}, &stubSearch{}, &stubExecute{})
	contents, err := h.HandleListAPIResource(context.Background(), mcp.ReadResourceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tc, ok := mcp.AsTextResourceContents(contents[0])
	if !ok {
		t.Fatal("expected TextResourceContents")
	}
	if !strings.Contains(tc.Text, "petstore") {
		t.Errorf("response should contain 'petstore', got: %q", tc.Text)
	}
}

// --- search ---

func TestHandleSearch_Success(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{result: `[{"operationId":"listPets"}]`}, &stubExecute{})
	req := makeRequest(map[string]any{"api_name": "petstore", "exec_code": `search("pets")`})
	result, err := h.HandleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", extractText(result))
	}
	if !strings.Contains(extractText(result), "listPets") {
		t.Errorf("result should contain 'listPets', got: %q", extractText(result))
	}
}

func TestHandleSearch_MissingAPIName(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{result: "ok"}, &stubExecute{})
	req := makeRequest(map[string]any{"exec_code": `search("pets")`})
	result, err := h.HandleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing api_name")
	}
}

func TestHandleSearch_MissingCode(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{result: "ok"}, &stubExecute{})
	req := makeRequest(map[string]any{"api_name": "petstore"})
	result, err := h.HandleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing exec_code")
	}
}

func TestHandleSearch_SandboxError(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{err: errors.New("js error")}, &stubExecute{})
	req := makeRequest(map[string]any{"api_name": "petstore", "exec_code": `{{{`})
	result, err := h.HandleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for sandbox error")
	}
}

// --- execute ---

func TestHandleExecute_Success(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{}, &stubExecute{result: `{"id":1}`})
	req := makeRequest(map[string]any{"api_name": "petstore", "exec_code": `httpGet("/pets/1")`})
	result, err := h.HandleExecute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %v", extractText(result))
	}
	if !strings.Contains(extractText(result), `{"id":1}`) {
		t.Errorf("result should contain JSON, got: %q", extractText(result))
	}
}

func TestHandleExecute_MissingAPIName(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{}, &stubExecute{result: "ok"})
	req := makeRequest(map[string]any{"exec_code": `httpGet("/")`})
	result, err := h.HandleExecute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing api_name")
	}
}

func TestHandleExecute_SandboxError(t *testing.T) {
	h := NewHandlers(&stubListApi{}, &stubSearch{}, &stubExecute{err: errors.New("timeout")})
	req := makeRequest(map[string]any{"api_name": "petstore", "exec_code": `while(true){}`})
	result, err := h.HandleExecute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for sandbox error")
	}
}
