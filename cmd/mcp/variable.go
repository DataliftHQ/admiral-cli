package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/resolve"
	sdkclient "go.admiral.io/sdk/client"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

// SetVariableInput defines the input schema for the admiral_set_variable tool.
type SetVariableInput struct {
	Key         string `json:"key" jsonschema:"Variable name (e.g. IMAGE_TAG, DB_URL)"`
	Value       string `json:"value" jsonschema:"Variable value"`
	App         string `json:"app,omitempty" jsonschema:"Application name (omit for global scope)"`
	Env         string `json:"env,omitempty" jsonschema:"Environment name (requires app)"`
	Sensitive   *bool  `json:"sensitive,omitempty" jsonschema:"Encrypt at rest and mask in responses"`
	Type        string `json:"type,omitempty" jsonschema:"Value type,enum=string,enum=number,enum=boolean,enum=complex"`
	Description string `json:"description,omitempty" jsonschema:"Purpose of this variable"`
}

// SetVariableOutput is the structured output from the set_variable tool.
type SetVariableOutput struct {
	Action   string               `json:"action"`
	Variable *variablev1.Variable `json:"variable"`
}

func handleSetVariable(c sdkclient.AdmiralClient) mcp.ToolHandlerFor[SetVariableInput, SetVariableOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SetVariableInput) (*mcp.CallToolResult, SetVariableOutput, error) {
		appID, envID, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, input.Env)
		if err != nil {
			return nil, SetVariableOutput{}, err
		}

		// Check if the variable already exists at this scope.
		existingID, _ := resolve.VariableByKey(ctx, c.Variable(), input.Key, appID, envID)

		if existingID != "" {
			return updateExistingVariable(ctx, c, existingID, input)
		}
		return createNewVariable(ctx, c, input, appID, envID)
	}
}

func createNewVariable(
	ctx context.Context,
	c sdkclient.AdmiralClient,
	input SetVariableInput,
	appID, envID string,
) (*mcp.CallToolResult, SetVariableOutput, error) {
	req := &variablev1.CreateVariableRequest{
		Key:   input.Key,
		Value: input.Value,
	}

	if appID != "" {
		req.ApplicationId = &appID
	}
	if envID != "" {
		req.EnvironmentId = &envID
	}
	if input.Sensitive != nil && *input.Sensitive {
		req.Sensitive = true
	}
	if input.Type != "" {
		vt, err := resolve.VariableType(input.Type)
		if err != nil {
			return nil, SetVariableOutput{}, err
		}
		req.Type = vt
	}
	if input.Description != "" {
		req.Description = input.Description
	}

	resp, err := c.Variable().CreateVariable(ctx, req)
	if err != nil {
		return nil, SetVariableOutput{}, err
	}

	return nil, SetVariableOutput{Action: "created", Variable: resp.Variable}, nil
}

func updateExistingVariable(
	ctx context.Context,
	c sdkclient.AdmiralClient,
	id string,
	input SetVariableInput,
) (*mcp.CallToolResult, SetVariableOutput, error) {
	variable := &variablev1.Variable{Id: id}
	var paths []string

	variable.Value = input.Value
	paths = append(paths, "value")

	if input.Sensitive != nil {
		variable.Sensitive = *input.Sensitive
		paths = append(paths, "sensitive")
	}
	if input.Type != "" {
		vt, err := resolve.VariableType(input.Type)
		if err != nil {
			return nil, SetVariableOutput{}, err
		}
		variable.Type = vt
		paths = append(paths, "type")
	}
	if input.Description != "" {
		variable.Description = input.Description
		paths = append(paths, "description")
	}

	resp, err := c.Variable().UpdateVariable(ctx, &variablev1.UpdateVariableRequest{
		Variable:   variable,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
	})
	if err != nil {
		return nil, SetVariableOutput{}, err
	}

	return nil, SetVariableOutput{Action: "updated", Variable: resp.Variable}, nil
}

// DeleteVariableInput defines the input schema for the admiral_delete_variable tool.
type DeleteVariableInput struct {
	Key string `json:"key,omitempty" jsonschema:"Variable key to delete"`
	App string `json:"app,omitempty" jsonschema:"Application name (omit for global scope)"`
	Env string `json:"env,omitempty" jsonschema:"Environment name (requires app)"`
	ID  string `json:"id,omitempty" jsonschema:"Variable UUID (bypasses key+scope lookup)"`
}

// DeleteVariableOutput is the structured output from the delete_variable tool.
type DeleteVariableOutput struct {
	Deleted bool   `json:"deleted"`
	Key     string `json:"key,omitempty"`
	ID      string `json:"id"`
}

func handleDeleteVariable(c sdkclient.AdmiralClient) mcp.ToolHandlerFor[DeleteVariableInput, DeleteVariableOutput] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input DeleteVariableInput) (*mcp.CallToolResult, DeleteVariableOutput, error) {
		id := input.ID

		if id == "" {
			if input.Key == "" {
				return nil, DeleteVariableOutput{}, fmt.Errorf("variable key or id is required")
			}

			appID, envID, err := resolve.ScopeIDs(ctx, c.Application(), c.Environment(), input.App, input.Env)
			if err != nil {
				return nil, DeleteVariableOutput{}, err
			}

			id, err = resolve.VariableByKey(ctx, c.Variable(), input.Key, appID, envID)
			if err != nil {
				return nil, DeleteVariableOutput{}, err
			}
		}

		_, err := c.Variable().DeleteVariable(ctx, &variablev1.DeleteVariableRequest{
			VariableId: id,
		})
		if err != nil {
			return nil, DeleteVariableOutput{}, err
		}

		return nil, DeleteVariableOutput{
			Deleted: true,
			Key:     input.Key,
			ID:      id,
		}, nil
	}
}
