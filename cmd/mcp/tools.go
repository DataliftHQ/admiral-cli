package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	sdkclient "go.admiral.io/sdk/client"
)

func registerTools(s *mcp.Server, c sdkclient.AdmiralClient) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "admiral_query",
		Description: "Query Admiral resources. Lists or gets applications, environments, " +
			"variables, clusters, cluster tokens, cluster status, workloads, and " +
			"personal access tokens. Returns JSON. " +
			"Scope for variables: omit app for global, provide app for app-scoped, " +
			"provide app+env for environment-scoped. " +
			"Cluster-scoped resources (cluster_token, cluster_status, workload) require " +
			"cluster name or cluster_id.",
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
