package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// BuildPrompts returns the MCP prompt definitions.
func BuildPrompts() []mcp.Prompt {
	searchAPI := mcp.NewPrompt("search_api",
		mcp.WithPromptDescription("Generate JavaScript code to search endpoints in an API"),
		mcp.WithArgument("api_name",
			mcp.ArgumentDescription("Name of the API to search"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("query",
			mcp.ArgumentDescription("Natural language description of what to search for"),
			mcp.RequiredArgument(),
		),
	)
	return []mcp.Prompt{searchAPI}
}

// HandleSearchPrompt generates a JS search code snippet for the given API and query.
func HandleSearchPrompt(req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	apiName := req.Params.Arguments["api_name"]
	query := req.Params.Arguments["query"]

	code := fmt.Sprintf(`// Search endpoints in API: %s
// Query: %s
const results = search(%q);
JSON.stringify(results, null, 2);`, apiName, query, query)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Search %q in API %q", query, apiName),
		Messages: []mcp.PromptMessage{
			{
				Role:    mcp.RoleUser,
				Content: mcp.NewTextContent(code),
			},
		},
	}, nil
}
