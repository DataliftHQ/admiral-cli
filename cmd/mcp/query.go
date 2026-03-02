package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go.admiral.io/cli/internal/resolve"
	sdkclient "go.admiral.io/sdk/client"
	applicationv1 "go.admiral.io/sdk/proto/admiral/api/application/v1"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

// QueryInput defines the input schema for the admiral_query tool.
type QueryInput struct {
	Resource string `json:"resource" jsonschema:"Resource type to query,enum=app,enum=env,enum=variable"`
	Action   string `json:"action,omitempty" jsonschema:"Query action (default: list),enum=list,enum=get"`
	App      string `json:"app,omitempty" jsonschema:"Application name"`
	Env      string `json:"env,omitempty" jsonschema:"Environment name (requires app)"`
	Key      string `json:"key,omitempty" jsonschema:"Variable key (for variable get)"`
	ID       string `json:"id,omitempty" jsonschema:"Resource UUID (bypasses name resolution)"`
	PageSize int32  `json:"page_size,omitempty" jsonschema:"Max results per page (default 50)"`
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
