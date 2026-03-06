package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	sdkclient "go.admiral.io/sdk/client"
)

func registerTools(s *mcp.Server, c sdkclient.AdmiralClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "admiral_query",
		Description: "Read-only query for Admiral resources. Set 'resource' and 'action' (list or get). " +
			"See server instructions for required fields per resource type. " +
			"All name fields are resolved to UUIDs automatically. Returns proto3 JSON (camelCase).",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, handleQuery(c))

	mcp.AddTool(s, &mcp.Tool{
		Name: "admiral_set_variable",
		Description: "Create or update a configuration variable. " +
			"If the variable already exists at the given scope, it is updated; otherwise created. " +
			"Scope is determined by which of app/env are provided: " +
			"neither = global, app only = app-scoped, app+env = environment-scoped.",
	}, handleSetVariable(c))

	mcp.AddTool(s, &mcp.Tool{
		Name: "admiral_delete_variable",
		Description: "Delete a configuration variable. This is destructive and cannot be undone. " +
			"Resolve by key+scope or provide the variable UUID directly via the id field.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
		},
	}, handleDeleteVariable(c))
}

func boolPtr(b bool) *bool { return &b }
