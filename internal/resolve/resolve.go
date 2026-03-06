package resolve

import (
	"context"
	"fmt"
	"strings"

	"go.admiral.io/cli/internal/cmdutil"
	applicationv1 "go.admiral.io/sdk/proto/admiral/api/application/v1"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

// ScopeIDs converts optional app/env names to UUIDs.
// Returns empty strings for global scope (no app provided).
func ScopeIDs(
	ctx context.Context,
	appClient applicationv1.ApplicationAPIClient,
	envClient environmentv1.EnvironmentAPIClient,
	app, env string,
) (appID, envID string, err error) {
	if app == "" {
		return "", "", nil
	}

	appID, err = cmdutil.ResolveAppID(ctx, appClient, app, "")
	if err != nil {
		return "", "", err
	}

	if env != "" {
		envID, err = cmdutil.ResolveEnvID(ctx, envClient, appID, env, "")
		if err != nil {
			return "", "", err
		}
	}

	return appID, envID, nil
}

// VariableByKey lists variables filtered by key and scope, returning the
// matching variable UUID. Returns an error if no variable or multiple variables match.
func VariableByKey(
	ctx context.Context,
	varClient variablev1.VariableAPIClient,
	key, appID, envID string,
) (string, error) {
	filters := []string{fmt.Sprintf("field['key'] = '%s'", key)}
	if appID != "" {
		filters = append(filters, fmt.Sprintf("field['application_id'] = '%s'", appID))
	}
	if envID != "" {
		filters = append(filters, fmt.Sprintf("field['environment_id'] = '%s'", envID))
	}

	resp, err := varClient.ListVariables(ctx, &variablev1.ListVariablesRequest{
		Filter: strings.Join(filters, " AND "),
	})
	if err != nil {
		return "", fmt.Errorf("looking up variable %q: %w", key, err)
	}

	// Client-side filter: match exact scope (API returns merged results).
	var matched []*variablev1.Variable
	for _, v := range resp.Variables {
		vAppID := ""
		if v.ApplicationId != nil {
			vAppID = *v.ApplicationId
		}
		vEnvID := ""
		if v.EnvironmentId != nil {
			vEnvID = *v.EnvironmentId
		}
		if vAppID == appID && vEnvID == envID {
			matched = append(matched, v)
		}
	}

	switch len(matched) {
	case 0:
		return "", fmt.Errorf("variable %q not found", key)
	case 1:
		return matched[0].Id, nil
	default:
		return "", fmt.Errorf("multiple variables match key %q; use --id to specify", key)
	}
}

// VariableFilter constructs a PEG filter string for ListVariables.
func VariableFilter(appID, envID string) string {
	var filters []string
	if appID != "" {
		filters = append(filters, fmt.Sprintf("field['application_id'] = '%s'", appID))
	}
	if envID != "" {
		filters = append(filters, fmt.Sprintf("field['environment_id'] = '%s'", envID))
	}
	return strings.Join(filters, " AND ")
}

// VariableType maps string representations to the proto VariableType enum.
// Supports aliases: str→string, num→number, bool→boolean. Case-insensitive.
func VariableType(s string) (variablev1.VariableType, error) {
	switch strings.ToLower(s) {
	case "string", "str":
		return variablev1.VariableType_VARIABLE_TYPE_STRING, nil
	case "number", "num":
		return variablev1.VariableType_VARIABLE_TYPE_NUMBER, nil
	case "boolean", "bool":
		return variablev1.VariableType_VARIABLE_TYPE_BOOLEAN, nil
	case "complex":
		return variablev1.VariableType_VARIABLE_TYPE_COMPLEX, nil
	default:
		return variablev1.VariableType_VARIABLE_TYPE_UNSPECIFIED,
			fmt.Errorf("unsupported variable type %q; valid values: string, number, boolean, complex", s)
	}
}
