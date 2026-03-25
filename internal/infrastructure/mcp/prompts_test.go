package mcp

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestBuildPrompts_SearchAPIExists(t *testing.T) {
	prompts := BuildPrompts()
	found := false
	for _, p := range prompts {
		if p.Name == "search_api" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("prompt 'search_api' not found")
	}
}

func TestBuildPrompts_SearchAPIHasArgs(t *testing.T) {
	prompts := BuildPrompts()
	var searchPrompt *mcp.Prompt
	for i := range prompts {
		if prompts[i].Name == "search_api" {
			searchPrompt = &prompts[i]
			break
		}
	}
	if searchPrompt == nil {
		t.Fatal("search_api prompt not found")
	}

	argNames := make(map[string]bool)
	for _, a := range searchPrompt.Arguments {
		argNames[a.Name] = true
	}
	if !argNames["api_name"] {
		t.Error("prompt should have 'api_name' argument")
	}
	if !argNames["query"] {
		t.Error("prompt should have 'query' argument")
	}
}

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
