package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/nuvivo314/mcp-executor/internal/config"
)

// NewMCPServer wires up the MCPServer with all tools, prompts, and resources.
func NewMCPServer(handlers *Handlers) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("mcp-executor", "1.0.0",
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithPromptCapabilities(true),
		mcpserver.WithResourceCapabilities(false, false),
	)
	registerTools(s, handlers)
	registerPrompts(s)
	registerResources(s, handlers)
	return s
}

// corsMiddleware adds permissive CORS headers so browser-based clients
// (e.g. MCP Inspector) can connect without a proxy.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id, Last-Event-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewStreamableServer creates the transport server from a wired MCPServer and config.
func NewStreamableServer(mcpSrv *mcpserver.MCPServer, cfg config.ServerConfig) (*mcpserver.StreamableHTTPServer, error) {
	if cfg.Transport != "streamable-http" && cfg.Transport != "sse" {
		return nil, fmt.Errorf("unsupported transport %q: use streamable-http or sse", cfg.Transport)
	}
	// Forward reference: srv is set before Start() accepts connections.
	var srv *mcpserver.StreamableHTTPServer
	mux := http.NewServeMux()
	mux.Handle("/mcp", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)
	})))
	srv = mcpserver.NewStreamableHTTPServer(mcpSrv,
		mcpserver.WithStreamableHTTPServer(&http.Server{Handler: mux}),
	)
	return srv, nil
}

// promptHandlerFunc is a prompt handler that does not need context.
type promptHandlerFunc func(mcp.GetPromptRequest) (*mcp.GetPromptResult, error)

// addPrompt registers a prompt on s, adapting the context-free handler signature.
func addPrompt(s *mcpserver.MCPServer, p mcp.Prompt, h promptHandlerFunc) {
	s.AddPrompt(p, func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return h(req)
	})
}
