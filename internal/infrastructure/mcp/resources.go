package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const listAPIResourceURI = "api://list"

// HandleListAPIResource handles reads of the api://list resource.
func (h *Handlers) HandleListAPIResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	apis := h.list.Execute()
	b, err := json.MarshalIndent(apis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialising APIs: %w", err)
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      listAPIResourceURI,
			MIMEType: "application/json",
			Text:     string(b),
		},
	}, nil
}

// registerResources adds all resource definitions and their handlers to s.
func registerResources(s *mcpserver.MCPServer, h *Handlers) {
	s.AddResource(
		mcp.NewResource(listAPIResourceURI, "list_api",
			mcp.WithResourceDescription("List all configured APIs available for search and execution"),
			mcp.WithMIMEType("application/json"),
		),
		h.HandleListAPIResource,
	)
}
