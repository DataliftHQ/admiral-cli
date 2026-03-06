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
	Resource  string `json:"resource" jsonschema:"Resource type to query,enum=app,enum=env,enum=variable,enum=cluster,enum=cluster_token,enum=cluster_status,enum=workload,enum=token,enum=whoami"`
	Action    string `json:"action,omitempty" jsonschema:"Query action (default: list),enum=list,enum=get"`
	App       string `json:"app,omitempty" jsonschema:"Application name"`
	Env       string `json:"env,omitempty" jsonschema:"Environment name (requires app)"`
	Cluster   string `json:"cluster,omitempty" jsonschema:"Cluster name (for cluster-scoped resources)"`
	ClusterID string `json:"cluster_id,omitempty" jsonschema:"Cluster UUID (bypasses cluster name resolution)"`
	Key       string `json:"key,omitempty" jsonschema:"Variable key (for variable get)"`
	ID        string `json:"id,omitempty" jsonschema:"Resource UUID (bypasses name resolution)"`
	Filter    string `json:"filter,omitempty" jsonschema:"Filter expression for list actions (see server instructions for syntax)"`
	PageSize  int32  `json:"page_size,omitempty" jsonschema:"Max results per page (default 50)"`
	PageToken string `json:"page_token,omitempty" jsonschema:"Pagination token from a previous list response"`
}

// QueryOutput is the structured output from the query tool.
type QueryOutput struct {
	Resource      string `json:"resource"`
	Action        string `json:"action"`
	Result        any    `json:"result"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

func handleQuery(c sdkclient.AdmiralClient) mcp.ToolHandlerFor[QueryInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, any, error) {
		action := input.Action
		if action == "" {
			action = "list"
		}

		pageSize := input.PageSize
		if pageSize == 0 {
			pageSize = 50
		}

		var (
			result        any
			nextPageToken string
			err           error
		)

		switch input.Resource + "." + action {
		case "variable.list":
			result, nextPageToken, err = queryVariableList(ctx, c, input, pageSize)
		case "variable.get":
			result, err = queryVariableGet(ctx, c, input)
		case "app.list":
			result, nextPageToken, err = queryAppList(ctx, c, input, pageSize)
		case "app.get":
			result, err = queryAppGet(ctx, c, input)
		case "env.list":
			result, nextPageToken, err = queryEnvList(ctx, c, input, pageSize)
		case "env.get":
			result, err = queryEnvGet(ctx, c, input)
		case "cluster.list":
			result, nextPageToken, err = queryClusterList(ctx, c, input, pageSize)
		case "cluster.get":
			result, err = queryClusterGet(ctx, c, input)
		case "cluster_status.get":
			result, err = queryClusterStatusGet(ctx, c, input)
		case "cluster_token.list":
			result, nextPageToken, err = queryClusterTokenList(ctx, c, input, pageSize)
		case "cluster_token.get":
			result, err = queryClusterTokenGet(ctx, c, input)
		case "workload.list":
			result, nextPageToken, err = queryWorkloadList(ctx, c, input, pageSize)
		case "token.list":
			result, nextPageToken, err = queryTokenList(ctx, c, input, pageSize)
		case "token.get":
			result, err = queryTokenGet(ctx, c, input)
		case "whoami.get":
			result, err = queryWhoami(ctx, c)
		default:
			return nil, nil, fmt.Errorf("unsupported query: %s.%s", input.Resource, action)
		}
		if err != nil {
			return nil, nil, err
		}

		return nil, QueryOutput{
			Resource:      input.Resource,
			Action:        action,
			Result:        result,
			NextPageToken: nextPageToken,
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

func queryVariableList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	appID, envID, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, input.Env)
	if err != nil {
		return nil, "", err
	}

	filter := resolve.VariableFilter(appID, envID)
	if input.Filter != "" {
		if filter != "" {
			filter += " AND " + input.Filter
		} else {
			filter = input.Filter
		}
	}

	resp, err := c.Variable().ListVariables(ctx, &variablev1.ListVariablesRequest{
		Filter:    filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
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

func queryAppList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	resp, err := c.Application().ListApplications(ctx, &applicationv1.ListApplicationsRequest{
		Filter:    input.Filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
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

func queryEnvList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	if input.App == "" {
		return nil, "", fmt.Errorf("app is required to list environments")
	}

	appID, _, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, "")
	if err != nil {
		return nil, "", err
	}

	filter := fmt.Sprintf("field['application_id'] = '%s'", appID)
	if input.Filter != "" {
		filter += " AND " + input.Filter
	}

	resp, err := c.Environment().ListEnvironments(ctx, &environmentv1.ListEnvironmentsRequest{
		Filter:    filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
}

// ---------------------------------------------------------------------------
// Cluster queries
// ---------------------------------------------------------------------------

func queryClusterList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	resp, err := c.Cluster().ListClusters(ctx, &clusterv1.ListClustersRequest{
		Filter:    input.Filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
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

func queryClusterTokenList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	clusterID, err := resolveClusterInput(ctx, c, input)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.Cluster().ListClusterTokens(ctx, &clusterv1.ListClusterTokensRequest{
		ClusterId: clusterID,
		Filter:    input.Filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
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

func queryWorkloadList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	clusterID, err := resolveClusterInput(ctx, c, input)
	if err != nil {
		return nil, "", err
	}

	resp, err := c.Cluster().ListWorkloads(ctx, &clusterv1.ListWorkloadsRequest{
		ClusterId: clusterID,
		Filter:    input.Filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
}

// ---------------------------------------------------------------------------
// Personal access token queries
// ---------------------------------------------------------------------------

func queryTokenList(ctx context.Context, c sdkclient.AdmiralClient, input QueryInput, pageSize int32) (any, string, error) {
	resp, err := c.User().ListPersonalAccessTokens(ctx, &userv1.ListPersonalAccessTokensRequest{
		Filter:    input.Filter,
		PageSize:  pageSize,
		PageToken: input.PageToken,
	})
	if err != nil {
		return nil, "", err
	}
	return resp, resp.NextPageToken, nil
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

// ---------------------------------------------------------------------------
// Whoami query
// ---------------------------------------------------------------------------

func queryWhoami(ctx context.Context, c sdkclient.AdmiralClient) (any, error) {
	resp, err := c.User().GetUser(ctx, &userv1.GetUserRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetUser(), nil
}
