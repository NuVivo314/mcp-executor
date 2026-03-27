package mcp

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleSearchPrompt_GeneratesJS(t *testing.T) {
	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"api_name": "petstore",
		"query":    "list pets",
	}
	result, err := HandleSearchPrompt(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "petstore") {
		t.Errorf("generated JS should reference api name, got: %q", text)
	}
	if !strings.Contains(text, "list pets") {
		t.Errorf("generated JS should reference query, got: %q", text)
	}
	if !strings.Contains(text, "search(") {
		t.Errorf("generated JS should call search(), got: %q", text)
	}
}

func TestHandleExplorePrompt_GeneratesJS(t *testing.T) {
	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{"api_name": "petstore"}
	result, err := HandleExplorePrompt(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "getEndpoints()") {
		t.Errorf("should call getEndpoints(), got: %q", text)
	}
	if !strings.Contains(text, "petstore") {
		t.Errorf("should reference api name, got: %q", text)
	}
}

func TestHandleExecuteGetPrompt_GeneratesJS(t *testing.T) {
	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"api_name": "petstore",
		"path":     "/pets/1",
	}
	result, err := HandleExecuteGetPrompt(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "httpGet(") {
		t.Errorf("should call httpGet(), got: %q", text)
	}
	if !strings.Contains(text, "/pets/1") {
		t.Errorf("should reference path, got: %q", text)
	}
}

func TestHandleExecutePostPrompt_GeneratesJS(t *testing.T) {
	req := mcp.GetPromptRequest{}
	req.Params.Arguments = map[string]string{
		"api_name": "petstore",
		"path":     "/pets",
		"body":     `{"name":"Rex"}`,
	}
	result, err := HandleExecutePostPrompt(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "httpPost(") {
		t.Errorf("should call httpPost(), got: %q", text)
	}
	if !strings.Contains(text, "/pets") {
		t.Errorf("should reference path, got: %q", text)
	}
	if !strings.Contains(text, "Rex") {
		t.Errorf("should include body, got: %q", text)
	}
}
