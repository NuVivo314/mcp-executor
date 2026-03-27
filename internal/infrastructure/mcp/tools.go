package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/nuvivo314/mcp-executor/internal/domain/model"
)

// listAPIUseCase is the interface expected by Handlers for list_api.
type listAPIUseCase interface {
	Execute() []model.APISpec
}

// searchUseCase is the interface expected by Handlers for search.
type searchUseCase interface {
	Execute(ctx context.Context, apiName, code string) (string, error)
}

// executeUseCase is the interface expected by Handlers for execute.
type executeUseCase interface {
	Execute(ctx context.Context, apiName, code string) (string, error)
}

// Handlers holds the use cases and exposes MCP tool and resource handler functions.
type Handlers struct {
	list    listAPIUseCase
	search  searchUseCase
	execute executeUseCase
}

// NewHandlers creates a Handlers with the given use cases.
func NewHandlers(list listAPIUseCase, search searchUseCase, execute executeUseCase) *Handlers {
	return &Handlers{list: list, search: search, execute: execute}
}

// handleSandbox extracts api_name + exec_code from the request and delegates to uc.
func handleSandbox(ctx context.Context, req mcp.CallToolRequest, name string, uc searchUseCase) (*mcp.CallToolResult, error) {
	apiName, err := req.RequireString("api_name")
	if err != nil {
		return mcp.NewToolResultError("api_name is required"), nil
	}
	code, err := req.RequireString("exec_code")
	if err != nil {
		return mcp.NewToolResultError("exec_code is required"), nil
	}
	slog.Debug("tool called", "tool", name, "api", apiName, "code_len", len(code))

	result, err := uc.Execute(ctx, apiName, code)
	if err != nil {
		slog.Debug(name+" error", "api", apiName, "err", err)
		return mcp.NewToolResultError(fmt.Sprintf("%s error: %v", name, err)), nil
	}
	slog.Debug(name+" result", "api", apiName, "result_len", len(result))
	return mcp.NewToolResultText(result), nil
}

// HandleSearch handles the search MCP tool call.
func (h *Handlers) HandleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleSandbox(ctx, req, "search", h.search)
}

// HandleExecute handles the execute MCP tool call.
func (h *Handlers) HandleExecute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return handleSandbox(ctx, req, "execute", h.execute)
}

// registerTools adds all tool definitions and their handlers to s.
func registerTools(s *mcpserver.MCPServer, h *Handlers) {
	s.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("Run JavaScript with search helpers (search, getEndpoints, getSpec) to explore an API"),
			mcp.WithString("api_name", mcp.Required(), mcp.Description("Name of the API to search")),
			mcp.WithString("exec_code", mcp.Required(), mcp.Description("JavaScript code to execute in the search context")),
		),
		h.HandleSearch,
	)
	s.AddTool(
		mcp.NewTool("execute",
			mcp.WithDescription("Run JavaScript with HTTP helpers (httpGet, httpPost, …) to call an API. Timeout: 60s"),
			mcp.WithString("api_name", mcp.Required(), mcp.Description("Name of the API to call")),
			mcp.WithString("exec_code", mcp.Required(), mcp.Description("JavaScript code to execute with HTTP helpers available")),
		),
		h.HandleExecute,
	)
}
