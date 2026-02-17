package app

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// AppCmd is the parent command for application operations.
type AppCmd struct {
	Cmd *cobra.Command
}

// NewAppCmd creates the app command tree.
func NewAppCmd(opts *factory.Options) *AppCmd {
	root := &AppCmd{}

	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage applications",
		Long: `Manage applications and their lifecycle.

Most commands accept an optional app argument. You can provide it in two ways:
  1. As a positional argument:  admiral app get my-api
  2. From the active context:   admiral use my-api && admiral app get`,
		Aliases:       []string{"application"},
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newListCmd(opts),
		newCreateCmd(opts),
		newGetCmd(opts),
		newUpdateCmd(opts),
		newDeleteCmd(opts),
		newStatusCmd(opts),
		newDiffCmd(opts),
		newCloneCmd(opts),
	)

	root.Cmd = cmd
	return root
}

// resolveApp determines the app slug from a positional argument or context.
func resolveApp(appArg, contextApp string) (string, error) {
	if appArg != "" {
		return appArg, nil
	}
	if contextApp != "" {
		return contextApp, nil
	}
	return "", fmt.Errorf("no app specified; use a positional argument or set context with 'admiral use <app>'")
}

// resolveAppWithHelp calls resolveApp and, on error, prints the
// command's help text before returning so the user sees usage context.
func resolveAppWithHelp(cmd *cobra.Command, appArg, contextApp string) (string, error) {
	app, err := resolveApp(appArg, contextApp)
	if err != nil {
		_ = cmd.Help()
		_, _ = fmt.Fprintln(cmd.ErrOrStderr())
		return "", err
	}
	return app, nil
}
