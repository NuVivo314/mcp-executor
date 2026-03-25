package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/nuvivo314/mcp-executor/internal/domain/model"
)

// listAPIUseCase is the interface expected by ToolHandlers for list_api.
type listAPIUseCase interface {
	Execute() []model.APISpec
}

// searchUseCase is the interface expected by ToolHandlers for search.
type searchUseCase interface {
	Execute(ctx context.Context, apiName, code string) (string, error)
}

// executeUseCase is the interface expected by ToolHandlers for execute.
type executeUseCase interface {
	Execute(ctx context.Context, apiName, code string) (string, error)
}

// ToolHandlers holds the use cases and exposes MCP tool handler functions.
type ToolHandlers struct {
	list    listAPIUseCase
	search  searchUseCase
	execute executeUseCase
}

// NewToolHandlers creates a ToolHandlers with the given use cases.
func NewToolHandlers(list listAPIUseCase, search searchUseCase, execute executeUseCase) *ToolHandlers {
	return &ToolHandlers{list: list, search: search, execute: execute}
}

// HandleListAPI handles the list_api MCP tool call.
func (h *ToolHandlers) HandleListAPI(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apis := h.list.Execute()
	b, err := json.MarshalIndent(apis, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("serialising APIs: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

// HandleSearch handles the search MCP tool call.
func (h *ToolHandlers) HandleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiName, err := req.RequireString("api_name")
	if err != nil {
		return mcp.NewToolResultError("api_name is required"), nil
	}
	code, err := req.RequireString("exec_code")
	if err != nil {
		return mcp.NewToolResultError("exec_code is required"), nil
	}

	result, err := h.search.Execute(ctx, apiName, code)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search error: %v", err)), nil
	}
	return mcp.NewToolResultText(result), nil
}

// HandleExecute handles the execute MCP tool call.
func (h *ToolHandlers) HandleExecute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiName, err := req.RequireString("api_name")
	if err != nil {
		return mcp.NewToolResultError("api_name is required"), nil
	}
	code, err := req.RequireString("exec_code")
	if err != nil {
		return mcp.NewToolResultError("exec_code is required"), nil
	}

	result, err := h.execute.Execute(ctx, apiName, code)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execute error: %v", err)), nil
	}
	return mcp.NewToolResultText(result), nil
}

// BuildTools returns the MCP tool definitions for list_api, search, execute.
func BuildTools() []mcp.Tool {
	listAPI := mcp.NewTool("list_api",
		mcp.WithDescription("List all configured APIs available for search and execution"),
	)

	search := mcp.NewTool("search",
		mcp.WithDescription("Run JavaScript with search helpers (search, getEndpoints, getSpec) to explore an API"),
		mcp.WithString("api_name",
			mcp.Required(),
			mcp.Description("Name of the API to search"),
		),
		mcp.WithString("exec_code",
			mcp.Required(),
			mcp.Description("JavaScript code to execute in the search context"),
		),
	)

	execute := mcp.NewTool("execute",
		mcp.WithDescription("Run JavaScript with HTTP helpers (httpGet, httpPost, …) to call an API. Timeout: 60s"),
		mcp.WithString("api_name",
			mcp.Required(),
			mcp.Description("Name of the API to call"),
		),
		mcp.WithString("exec_code",
			mcp.Required(),
			mcp.Description("JavaScript code to execute with HTTP helpers available"),
		),
	)

	return []mcp.Tool{listAPI, search, execute}
}
