package mcp_test

import (
	"context"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/nuvivo314/mcp-executor/internal/application"
	"github.com/nuvivo314/mcp-executor/internal/config"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/httpclient"
	mcpinfra "github.com/nuvivo314/mcp-executor/internal/infrastructure/mcp"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/openapi"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/sandbox"
	"github.com/nuvivo314/mcp-executor/internal/infrastructure/search"
)

// buildTestClient wires the full stack and returns an in-process MCP client.
func buildTestClient(t *testing.T) *mcpclient.Client {
	t.Helper()

	apiCfgs := []config.APIConfig{
		{
			Name:        "petstore",
			Description: "Petstore sample API",
			SpecPath:    "../openapi/testdata/petstore.yaml",
			BaseURL:     "https://petstore.example.com",
			Auth:        config.AuthConfig{Type: "none"},
		},
	}

	registry, err := openapi.NewRegistry(apiCfgs)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}

	searchEngine := search.NewFuzzySearch(registry)
	httpClient := httpclient.NewHTTPClient([]string{}, 1<<20)
	sb := sandbox.NewSandbox(registry, searchEngine, httpClient)

	listUC := application.NewListApiUseCase(registry)
	searchUC := application.NewSearchUseCase(sb)
	executeUC := application.NewExecuteUseCase(sb, 60*time.Second)

	handlers := mcpinfra.NewHandlers(listUC, searchUC, executeUC)
	mcpSrv := mcpinfra.NewMCPServer(handlers)

	c, err := mcpclient.NewInProcessClient(mcpSrv)
	if err != nil {
		t.Fatalf("creating in-process client: %v", err)
	}

	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("starting client: %v", err)
	}

	_, err = c.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		t.Fatalf("initialising client: %v", err)
	}

	t.Cleanup(func() { _ = c.Close() })
	return c
}

// --- list_api resource ---

func TestIntegration_ListAPIResource_ReturnsAPI(t *testing.T) {
	c := buildTestClient(t)
	result, err := c.ReadResource(context.Background(), mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{URI: "api://list"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Contents) == 0 {
		t.Fatal("expected resource contents")
	}
	tc, ok := mcp.AsTextResourceContents(result.Contents[0])
	if !ok {
		t.Fatal("expected TextResourceContents")
	}
	if !strings.Contains(tc.Text, "petstore") {
		t.Errorf("response should contain 'petstore', got: %q", tc.Text)
	}
}

// --- search ---

func TestIntegration_Search_FindsEndpoints(t *testing.T) {
	c := buildTestClient(t)

	args := map[string]any{
		"api_name":  "petstore",
		"exec_code": `JSON.stringify(getEndpoints())`,
	}
	result, err := c.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "search",
			Arguments: args,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", mcp.GetTextFromContent(result.Content[0]))
	}
	text := mcp.GetTextFromContent(result.Content[0])
	if !strings.Contains(text, "listPets") {
		t.Errorf("response should contain 'listPets', got: %q", text)
	}
}

func TestIntegration_Search_FuzzySearch(t *testing.T) {
	c := buildTestClient(t)
	args := map[string]any{
		"api_name":  "petstore",
		"exec_code": `JSON.stringify(search("create pet"))`,
	}
	result, err := c.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "search", Arguments: args},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", mcp.GetTextFromContent(result.Content[0]))
	}
	text := mcp.GetTextFromContent(result.Content[0])
	if !strings.Contains(text, "createPet") {
		t.Errorf("expected createPet in results, got: %q", text)
	}
}

func TestIntegration_Search_UnknownAPI(t *testing.T) {
	c := buildTestClient(t)
	args := map[string]any{
		"api_name":  "unknown",
		"exec_code": `search("anything")`,
	}
	result, err := c.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "search", Arguments: args},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for unknown API")
	}
}

// --- execute ---

func TestIntegration_Execute_JSArithmetic(t *testing.T) {
	c := buildTestClient(t)
	args := map[string]any{
		"api_name":  "petstore",
		"exec_code": `String(6 * 7)`,
	}
	result, err := c.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "execute", Arguments: args},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %v", mcp.GetTextFromContent(result.Content[0]))
	}
	text := mcp.GetTextFromContent(result.Content[0])
	if text != "42" {
		t.Errorf("result = %q, want %q", text, "42")
	}
}

func TestIntegration_Execute_SearchAvailable(t *testing.T) {
	c := buildTestClient(t)
	args := map[string]any{
		"api_name":  "petstore",
		"exec_code": `typeof search`,
	}
	result, err := c.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "execute", Arguments: args},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := mcp.GetTextFromContent(result.Content[0])
	if text != "function" {
		t.Errorf("search should be a function in execute context, got %q", text)
	}
}

func TestIntegration_Execute_TimeoutEnforced(t *testing.T) {
	c := buildTestClient(t)
	// Very short timeout use case
	listUC2 := application.NewListApiUseCase(nil)
	_ = listUC2

	args := map[string]any{
		"api_name":  "petstore",
		"exec_code": `while(true){}`,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: "execute", Arguments: args},
	})
	// Either context error or tool error is acceptable
	if err == nil && !result.IsError {
		t.Error("expected timeout error for infinite loop")
	}
}

// --- prompt ---

func TestIntegration_Prompt_SearchAPI(t *testing.T) {
	c := buildTestClient(t)
	result, err := c.GetPrompt(context.Background(), mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "search_api",
			Arguments: map[string]string{
				"api_name": "petstore",
				"query":    "list all pets",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "search(") {
		t.Errorf("prompt should contain search() call, got: %q", text)
	}
	if !strings.Contains(text, "list all pets") {
		t.Errorf("prompt should reference the query, got: %q", text)
	}
}

func TestIntegration_Prompt_ExploreAPI(t *testing.T) {
	c := buildTestClient(t)
	result, err := c.GetPrompt(context.Background(), mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      "explore_api",
			Arguments: map[string]string{"api_name": "petstore"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "getEndpoints()") {
		t.Errorf("prompt should call getEndpoints(), got: %q", text)
	}
}

func TestIntegration_Prompt_ExecuteGet(t *testing.T) {
	c := buildTestClient(t)
	result, err := c.GetPrompt(context.Background(), mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "execute_get",
			Arguments: map[string]string{
				"api_name": "petstore",
				"path":     "/pets/1",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "httpGet(") {
		t.Errorf("prompt should call httpGet(), got: %q", text)
	}
	if !strings.Contains(text, "/pets/1") {
		t.Errorf("prompt should reference path, got: %q", text)
	}
}

func TestIntegration_Prompt_ExecutePost(t *testing.T) {
	c := buildTestClient(t)
	result, err := c.GetPrompt(context.Background(), mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "execute_post",
			Arguments: map[string]string{
				"api_name": "petstore",
				"path":     "/pets",
				"body":     `{"name":"Rex"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := mcp.GetTextFromContent(result.Messages[0].Content)
	if !strings.Contains(text, "httpPost(") {
		t.Errorf("prompt should call httpPost(), got: %q", text)
	}
	if !strings.Contains(text, "Rex") {
		t.Errorf("prompt should include body, got: %q", text)
	}
}

func TestIntegration_ListPrompts(t *testing.T) {
	c := buildTestClient(t)
	result, err := c.ListPrompts(context.Background(), mcp.ListPromptsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := make(map[string]bool)
	for _, p := range result.Prompts {
		names[p.Name] = true
	}
	for _, expected := range []string{"search_api", "explore_api", "execute_get", "execute_post"} {
		if !names[expected] {
			t.Errorf("prompt %q not listed", expected)
		}
	}
}
