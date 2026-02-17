package variable

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"go.admiral.io/cli/internal/output"
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
//   - no app + no --global = error
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
		return resolvedScope{}, fmt.Errorf("no app specified; use a positional argument or set context with 'admiral use <app>'")
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

// splitArgs separates mixed positional arguments for the set command.
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
			return "", "", fmt.Errorf("unexpected KEY=VALUE argument %q; use 'variable set' to set values", arg)
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

// stubResult holds the parameters that would be sent to the API.
type stubResult struct {
	Operation   string            `json:"operation" yaml:"operation"`
	Scope       string            `json:"scope" yaml:"scope"`
	App         string            `json:"app,omitempty" yaml:"app,omitempty"`
	Environment string            `json:"environment,omitempty" yaml:"environment,omitempty"`
	Key         string            `json:"key,omitempty" yaml:"key,omitempty"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Sensitive   bool              `json:"sensitive,omitempty" yaml:"sensitive,omitempty"`
	Status      string            `json:"status" yaml:"status"`
}

const stubStatus = "not yet implemented (proto/gRPC pending)"

func printStub(w io.Writer, format output.Format, stub stubResult) error {
	switch format {
	case output.FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stub)
	case output.FormatYAML:
		return yaml.NewEncoder(w).Encode(stub)
	case output.FormatTable, output.FormatWide:
		return printStubTable(w, stub)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func printStubTable(out io.Writer, stub stubResult) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)

	output.Writef(w, "Operation:\t%s\n", stub.Operation)
	output.Writef(w, "Scope:\t%s\n", stub.Scope)

	if stub.App != "" {
		output.Writef(w, "App:\t%s\n", stub.App)
	}
	if stub.Environment != "" {
		output.Writef(w, "Environment:\t%s\n", stub.Environment)
	}

	if stub.Key != "" {
		output.Writef(w, "Key:\t%s\n", stub.Key)
	}

	if len(stub.Variables) > 0 {
		output.Writeln(w, "Variables:")
		for k, v := range stub.Variables {
			output.Writef(w, "  %s\t= %s\n", k, v)
		}
	}

	if stub.Sensitive {
		output.Writef(w, "Sensitive:\ttrue\n")
	}

	output.Writef(w, "\nStatus:\t%s\n", stub.Status)

	return w.Flush()
}
