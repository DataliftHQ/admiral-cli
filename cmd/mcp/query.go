package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/resolve"
	sdkclient "go.admiral.io/sdk/client"
	applicationv1 "go.admiral.io/sdk/proto/admiral/api/application/v1"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
	userv1 "go.admiral.io/sdk/proto/admiral/api/user/v1"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

// QueryInput defines the input schema for the admiral_query tool.
type QueryInput struct {
	Resource  string `json:"resource" jsonschema:"Resource type to query,enum=app,enum=env,enum=variable,enum=cluster,enum=cluster_token,enum=cluster_status,enum=workload,enum=token"`
	Action    string `json:"action,omitempty" jsonschema:"Query action (default: list),enum=list,enum=get"`
	App       string `json:"app,omitempty" jsonschema:"Application name"`
	Env       string `json:"env,omitempty" jsonschema:"Environment name (requires app)"`
	Cluster   string `json:"cluster,omitempty" jsonschema:"Cluster name (for cluster_token, cluster_status, workload queries)"`
	ClusterID string `json:"cluster_id,omitempty" jsonschema:"Cluster UUID (bypasses cluster name resolution)"`
	Key       string `json:"key,omitempty" jsonschema:"Variable key (for variable get)"`
	ID        string `json:"id,omitempty" jsonschema:"Resource UUID (bypasses name resolution)"`
	PageSize  int32  `json:"page_size,omitempty" jsonschema:"Max results per page (default 50)"`
}

// QueryOutput is the structured output from the query tool.
type QueryOutput struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Result   any    `json:"result"`
}

func handleQuery(c sdkclient.AdmiralClient) mcp.ToolHandlerFor[QueryInput, QueryOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, QueryOutput, error) {
		action := input.Action
		if action == "" {
			action = "list"
		}

		pageSize := input.PageSize
		if pageSize == 0 {
			pageSize = 50
		}

		var result any
		var err error

		switch input.Resource + "." + action {
		case "variable.list":
			result, err = queryVariableList(ctx, c, input, pageSize)
		case "variable.get":
			result, err = queryVariableGet(ctx, c, input)
		case "app.list":
			result, err = queryAppList(ctx, c, pageSize)
		case "app.get":
			result, err = queryAppGet(ctx, c, input)
		case "env.list":
			result, err = queryEnvList(ctx, c, input, pageSize)
		case "env.get":
			result, err = queryEnvGet(ctx, c, input)
		case "cluster.list":
			result, err = queryClusterList(ctx, c, pageSize)
		case "cluster.get":
			result, err = queryClusterGet(ctx, c, input)
		case "cluster_status.get":
			result, err = queryClusterStatusGet(ctx, c, input)
		case "cluster_token.list":
			result, err = queryClusterTokenList(ctx, c, input, pageSize)
		case "cluster_token.get":
			result, err = queryClusterTokenGet(ctx, c, input)
		case "workload.list":
			result, err = queryWorkloadList(ctx, c, input, pageSize)
		case "token.list":
			result, err = queryTokenList(ctx, c, pageSize)
		case "token.get":
			result, err = queryTokenGet(ctx, c, input)
		default:
			return nil, QueryOutput{}, fmt.Errorf("unsupported query: %s.%s", input.Resource, action)
		}
		if err != nil {
			return nil, QueryOutput{}, err
		}

		return nil, QueryOutput{
			Resource: input.Resource,
			Action:   action,
			Result:   result,
		}, nil
	}
}

// ---------------------------------------------------------------------------
// Cluster helpers
// ---------------------------------------------------------------------------

// resolveClusterInput resolves cluster name/ID from query input fields.
func resolveClusterInput(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (string, error) {
	if input.ClusterID != "" {
		return input.ClusterID, nil
	}
	if input.Cluster == "" {
		return "", fmt.Errorf("cluster name or cluster_id is required")
	}
	return cmdutil.ResolveClusterID(ctx, c.Cluster(), input.Cluster, "")
}

// ---------------------------------------------------------------------------
// Variable queries
// ---------------------------------------------------------------------------

func queryVariableList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, error) {
	appID, envID, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, input.Env)
	if err != nil {
		return nil, err
	}

	resp, err := c.Variable().ListVariables(ctx, &variablev1.ListVariablesRequest{
		Filter:   resolve.VariableFilter(appID, envID),
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func queryVariableGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	id := input.ID
	if id == "" {
		if input.Key == "" {
			return nil, fmt.Errorf("variable key or id is required")
		}
		appID, envID, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, input.Env)
		if err != nil {
			return nil, err
		}
		id, err = resolve.VariableByKey(ctx, c.Variable(), input.Key, appID, envID)
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.Variable().GetVariable(ctx, &variablev1.GetVariableRequest{
		VariableId: id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Variable, nil
}

// ---------------------------------------------------------------------------
// App queries
// ---------------------------------------------------------------------------

func queryAppList(ctx context.Context, c sdkclient.AdmiralClient, pageSize int32) (any, error) {
	resp, err := c.Application().ListApplications(ctx, &applicationv1.ListApplicationsRequest{
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func queryAppGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	if input.ID == "" && input.App == "" {
		return nil, fmt.Errorf("app name or id is required")
	}

	id := input.ID
	if id == "" {
		var err error
		id, _, err = resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, "")
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.Application().GetApplication(ctx, &applicationv1.GetApplicationRequest{
		ApplicationId: id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Application, nil
}

// ---------------------------------------------------------------------------
// Env queries
// ---------------------------------------------------------------------------

func queryEnvGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	if input.ID == "" && (input.App == "" || input.Env == "") {
		return nil, fmt.Errorf("app and env names (or id) are required")
	}

	id := input.ID
	if id == "" {
		appID, envID, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, input.Env)
		if err != nil {
			return nil, err
		}
		id = envID
		_ = appID // consumed by ScopeIDs to resolve envID
	}

	resp, err := c.Environment().GetEnvironment(ctx, &environmentv1.GetEnvironmentRequest{
		EnvironmentId: id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Environment, nil
}

func queryEnvList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, error) {
	if input.App == "" {
		return nil, fmt.Errorf("app is required to list environments")
	}

	appID, _, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, "")
	if err != nil {
		return nil, err
	}

	resp, err := c.Environment().ListEnvironments(ctx, &environmentv1.ListEnvironmentsRequest{
		Filter:   fmt.Sprintf("field['application_id'] = '%s'", appID),
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ---------------------------------------------------------------------------
// Cluster queries
// ---------------------------------------------------------------------------

func queryClusterList(ctx context.Context, c sdkclient.AdmiralClient, pageSize int32) (any, error) {
	resp, err := c.Cluster().ListClusters(ctx, &clusterv1.ListClustersRequest{
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func queryClusterGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	if input.ID == "" && input.Cluster == "" && input.ClusterID == "" {
		return nil, fmt.Errorf("cluster name, cluster_id, or id is required")
	}

	id := input.ID
	if id == "" {
		var err error
		id, err = resolveClusterInput(ctx, c, input)
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.Cluster().GetCluster(ctx, &clusterv1.GetClusterRequest{
		ClusterId: id,
	})
	if err != nil {
		return nil, err
	}
	return resp.Cluster, nil
}

func queryClusterStatusGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	if input.ID == "" && input.Cluster == "" && input.ClusterID == "" {
		return nil, fmt.Errorf("cluster name, cluster_id, or id is required")
	}

	id := input.ID
	if id == "" {
		var err error
		id, err = resolveClusterInput(ctx, c, input)
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.Cluster().GetClusterStatus(ctx, &clusterv1.GetClusterStatusRequest{
		ClusterId: id,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ---------------------------------------------------------------------------
// Cluster token queries
// ---------------------------------------------------------------------------

func queryClusterTokenList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, error) {
	clusterID, err := resolveClusterInput(ctx, c, input)
	if err != nil {
		return nil, err
	}

	resp, err := c.Cluster().ListClusterTokens(ctx, &clusterv1.ListClusterTokensRequest{
		ClusterId: clusterID,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func queryClusterTokenGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	clusterID, err := resolveClusterInput(ctx, c, input)
	if err != nil {
		return nil, err
	}

	tokenID := input.ID
	if tokenID == "" {
		return nil, fmt.Errorf("token id is required (token names may not be unique)")
	}

	resp, err := c.Cluster().GetClusterToken(ctx, &clusterv1.GetClusterTokenRequest{
		ClusterId: clusterID,
		TokenId:   tokenID,
	})
	if err != nil {
		return nil, err
	}
	return resp.AccessToken, nil
}

// ---------------------------------------------------------------------------
// Workload queries
// ---------------------------------------------------------------------------

func queryWorkloadList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, error) {
	clusterID, err := resolveClusterInput(ctx, c, input)
	if err != nil {
		return nil, err
	}

	resp, err := c.Cluster().ListWorkloads(ctx, &clusterv1.ListWorkloadsRequest{
		ClusterId: clusterID,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ---------------------------------------------------------------------------
// Personal access token queries
// ---------------------------------------------------------------------------

func queryTokenList(ctx context.Context, c sdkclient.AdmiralClient, pageSize int32) (any, error) {
	resp, err := c.User().ListPersonalAccessTokens(ctx, &userv1.ListPersonalAccessTokensRequest{
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func queryTokenGet(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput) (any, error) {
	if input.ID == "" {
		return nil, fmt.Errorf("token id is required")
	}

	resp, err := c.User().GetPersonalAccessToken(ctx, &userv1.GetPersonalAccessTokenRequest{
		TokenId: input.ID,
	})
	if err != nil {
		return nil, err
	}
	return resp.AccessToken, nil
}
