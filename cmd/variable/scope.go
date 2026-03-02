package variable

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/output"
	variablev1 "go.admiral.io/sdk/proto/admiral/api/variable/v1"
)

// scope represents the resolved variable scope.
type scope string

const (
	scopeGlobal scope = "GLOBAL"
	scopeApp    scope = "APP"
	scopeAppEnv scope = "APP_ENV"
)

// resolvedScope holds the fully resolved scope with optional app and env.
type resolvedScope struct {
	Scope scope
	App   string
	Env   string
}

// resolveScope determines the variable scope from flags and context.
//
// Rules:
//   - --global is mutually exclusive with --env and app arg
//   - app arg overrides context app
//   - no app + no --global = GLOBAL (implicit)
//   - --env requires an app (from arg or context)
func resolveScope(globalFlag bool, envFlag string, appArg string, contextApp string) (resolvedScope, error) {
	app := contextApp
	if appArg != "" {
		app = appArg
	}

	if globalFlag {
		if envFlag != "" {
			return resolvedScope{}, fmt.Errorf("--global and --env are mutually exclusive")
		}
		if appArg != "" {
			return resolvedScope{}, fmt.Errorf("--global and app argument are mutually exclusive")
		}
		return resolvedScope{Scope: scopeGlobal}, nil
	}

	if app == "" {
		return resolvedScope{Scope: scopeGlobal}, nil
	}

	if envFlag != "" {
		return resolvedScope{Scope: scopeAppEnv, App: app, Env: envFlag}, nil
	}

	return resolvedScope{Scope: scopeApp, App: app}, nil
}

// resolveScopeWithHelp calls resolveScope and, on error, prints the
// command's help text before returning so the user sees usage context.
func resolveScopeWithHelp(cmd *cobra.Command, globalFlag bool, envFlag, appArg, contextApp string) (resolvedScope, error) {
	rs, err := resolveScope(globalFlag, envFlag, appArg, contextApp)
	if err != nil {
		_ = cmd.Help()
		_, _ = fmt.Fprintln(cmd.ErrOrStderr())
		return rs, err
	}
	return rs, nil
}

// splitArgs separates mixed positional arguments for the create command.
// Arguments containing "=" are treated as KEY=VALUE pairs.
// The first argument without "=" is treated as the app name.
// A second bare argument is an error.
func splitArgs(args []string) (appArg string, kvPairs []string, err error) {
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			kvPairs = append(kvPairs, arg)
		} else {
			if appArg != "" {
				return "", nil, fmt.Errorf("unexpected argument %q: app already set to %q", arg, appArg)
			}
			appArg = arg
		}
	}
	return appArg, kvPairs, nil
}

// parseKV parses a "KEY=VALUE" string into its components.
func parseKV(s string) (key, value string, err error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid variable format %q: expected KEY=VALUE", s)
	}
	return parts[0], parts[1], nil
}

// splitArgsWithKey separates positional arguments for get/delete commands.
// Expects exactly one bare arg (key) or two bare args (app + key).
// Arguments containing "=" are rejected.
func splitArgsWithKey(args []string) (appArg, key string, err error) {
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			return "", "", fmt.Errorf("unexpected KEY=VALUE argument %q; use 'variable create' to create variables", arg)
		}
	}

	switch len(args) {
	case 1:
		return "", args[0], nil
	case 2:
		return args[0], args[1], nil
	default:
		return "", "", fmt.Errorf("expected 1-2 arguments (KEY or APP KEY), got %d", len(args))
	}
}

// formatValue returns the variable value or a masked placeholder if sensitive.
func formatValue(v *variablev1.Variable) string {
	if v.Sensitive {
		return "********"
	}
	return v.Value
}

// formatScope returns a human-readable scope string based on the variable's IDs.
func formatScope(v *variablev1.Variable) string {
	if v.ApplicationId != nil && v.EnvironmentId != nil {
		return "App+Env"
	}
	if v.ApplicationId != nil {
		return "App"
	}
	return "Global"
}

// formatSensitive returns "Yes" or "No".
func formatSensitive(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// stringPtrOrNone returns the pointed-to string or "<none>".
func stringPtrOrNone(s *string) string {
	if s != nil {
		return *s
	}
	return "<none>"
}

// checkShadowWarning prints a best-effort warning to stderr if the key already
// exists at a broader scope that would be shadowed by the new variable.
func checkShadowWarning(
	ctx context.Context,
	stderr io.Writer,
	varClient variablev1.VariableAPIClient,
	key string,
	rs resolvedScope,
	appID, envID string,
) {
	// Only check for shadowing when creating at app or app+env scope.
	if rs.Scope == scopeGlobal {
		return
	}

	filters := []string{fmt.Sprintf("field['key'] = '%s'", key)}

	// For app+env scope, check if the key exists at app or global scope.
	// For app scope, check if the key exists at global scope.
	// We list with the app filter (which returns merged global+app results).
	if appID != "" {
		filters = append(filters, fmt.Sprintf("field['application_id'] = '%s'", appID))
	}

	resp, err := varClient.ListVariables(ctx, &variablev1.ListVariablesRequest{
		Filter: strings.Join(filters, " AND "),
	})
	if err != nil {
		return // best-effort, ignore errors
	}

	for _, v := range resp.Variables {
		vScope := formatScope(v)
		// Skip if this is the same scope we're creating at.
		vEnvID := ""
		if v.EnvironmentId != nil {
			vEnvID = *v.EnvironmentId
		}
		vAppID := ""
		if v.ApplicationId != nil {
			vAppID = *v.ApplicationId
		}
		if vAppID == appID && vEnvID == envID {
			continue
		}
		output.Writef(stderr, "Warning: %q already exists at %s scope and will be shadowed\n", key, vScope)
	}
}
