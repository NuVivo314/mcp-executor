package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/nuvivo314/mcp-executor/internal/config"
)

// NewMCPServer wires up the MCPServer with all tools and prompts.
func NewMCPServer(handlers *ToolHandlers) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("mcp-executor", "1.0.0",
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithPromptCapabilities(true),
	)

	for _, tool := range BuildTools() {
		t := tool // capture
		switch t.Name {
		case "list_api":
			s.AddTool(t, handlers.HandleListAPI)
		case "search":
			s.AddTool(t, handlers.HandleSearch)
		case "execute":
			s.AddTool(t, handlers.HandleExecute)
		}
	}

	for _, prompt := range BuildPrompts() {
		p := prompt // capture
		s.AddPrompt(p, func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return HandleSearchPrompt(req)
		})
	}

	return s
}

// NewStreamableServer creates the transport server from a wired MCPServer and config.
func NewStreamableServer(mcpSrv *mcpserver.MCPServer, cfg config.ServerConfig) (*mcpserver.StreamableHTTPServer, error) {
	if cfg.Transport != "streamable-http" && cfg.Transport != "sse" {
		return nil, fmt.Errorf("unsupported transport %q: use streamable-http or sse", cfg.Transport)
	}
	return mcpserver.NewStreamableHTTPServer(mcpSrv), nil
}
