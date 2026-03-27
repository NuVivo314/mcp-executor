package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// registerPrompts adds all prompt definitions and their handlers to s.
func registerPrompts(s *mcpserver.MCPServer) {
	addPrompt(s,
		mcp.NewPrompt("search_api",
			mcp.WithPromptDescription("Generate JavaScript code to search endpoints in an API"),
			mcp.WithArgument("api_name", mcp.ArgumentDescription("Name of the API to search"), mcp.RequiredArgument()),
			mcp.WithArgument("query", mcp.ArgumentDescription("Natural language description of what to search for"), mcp.RequiredArgument()),
		),
		HandleSearchPrompt,
	)
	addPrompt(s,
		mcp.NewPrompt("explore_api",
			mcp.WithPromptDescription("Generate JavaScript code to list all endpoints of an API"),
			mcp.WithArgument("api_name", mcp.ArgumentDescription("Name of the API to explore"), mcp.RequiredArgument()),
		),
		HandleExplorePrompt,
	)
	addPrompt(s,
		mcp.NewPrompt("execute_get",
			mcp.WithPromptDescription("Generate JavaScript code to perform a GET request on an API endpoint"),
			mcp.WithArgument("api_name", mcp.ArgumentDescription("Name of the API to call"), mcp.RequiredArgument()),
			mcp.WithArgument("path", mcp.ArgumentDescription("Endpoint path, e.g. /pets/1"), mcp.RequiredArgument()),
		),
		HandleExecuteGetPrompt,
	)
	addPrompt(s,
		mcp.NewPrompt("execute_post",
			mcp.WithPromptDescription("Generate JavaScript code to perform a POST request on an API endpoint"),
			mcp.WithArgument("api_name", mcp.ArgumentDescription("Name of the API to call"), mcp.RequiredArgument()),
			mcp.WithArgument("path", mcp.ArgumentDescription("Endpoint path, e.g. /pets"), mcp.RequiredArgument()),
			mcp.WithArgument("body", mcp.ArgumentDescription(`JSON body to send, e.g. {"name":"Rex"}`), mcp.RequiredArgument()),
		),
		HandleExecutePostPrompt,
	)
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
			{Role: mcp.RoleUser, Content: mcp.NewTextContent(code)},
		},
	}, nil
}

// HandleExplorePrompt generates JS to list all endpoints of an API.
func HandleExplorePrompt(req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	apiName := req.Params.Arguments["api_name"]

	code := fmt.Sprintf(`// Explore all endpoints of API: %s
const endpoints = getEndpoints();
JSON.stringify(endpoints, null, 2);`, apiName)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Explore all endpoints of API %q", apiName),
		Messages: []mcp.PromptMessage{
			{Role: mcp.RoleUser, Content: mcp.NewTextContent(code)},
		},
	}, nil
}

// HandleExecuteGetPrompt generates JS to perform a GET request.
func HandleExecuteGetPrompt(req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	apiName := req.Params.Arguments["api_name"]
	path := req.Params.Arguments["path"]

	code := fmt.Sprintf(`// GET %s on API: %s
const response = httpGet(%q);
JSON.stringify(response, null, 2);`, path, apiName, path)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("GET %s on API %q", path, apiName),
		Messages: []mcp.PromptMessage{
			{Role: mcp.RoleUser, Content: mcp.NewTextContent(code)},
		},
	}, nil
}

// HandleExecutePostPrompt generates JS to perform a POST request with a body.
func HandleExecutePostPrompt(req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	apiName := req.Params.Arguments["api_name"]
	path := req.Params.Arguments["path"]
	body := req.Params.Arguments["body"]

	code := fmt.Sprintf(`// POST %s on API: %s
const body = %s;
const response = httpPost(%q, JSON.stringify(body));
JSON.stringify(response, null, 2);`, path, apiName, body, path)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("POST %s on API %q", path, apiName),
		Messages: []mcp.PromptMessage{
			{Role: mcp.RoleUser, Content: mcp.NewTextContent(code)},
		},
	}, nil
}
