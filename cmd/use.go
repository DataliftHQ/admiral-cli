package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/properties"
)

func newUseCmd(opts *factory.Options) *cobra.Command {
	var clear bool

	cmd := &cobra.Command{
		Use:   "use [app-name]",
		Short: "Set the active app context",
		Long: `Set the active app context for subsequent commands.

When an app context is set, app-scoped commands use it implicitly.
CI/CD scripts should always pass --app explicitly instead.

  admiral use my-app       Set the active app
  admiral use              Show the current context
  admiral use --clear      Clear the active context`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if clear {
				if err := properties.Clear(opts.ConfigDir); err != nil {
					return fmt.Errorf("failed to clear context: %w", err)
				}
				output.Writef(cmd.OutOrStdout(), "Context cleared.\n")
				return nil
			}

			if len(args) == 1 {
				ctx := &properties.Properties{App: args[0]}
				if err := properties.Save(opts.ConfigDir, ctx); err != nil {
					return fmt.Errorf("failed to save context: %w", err)
				}
				output.Writef(cmd.OutOrStdout(), "Active app set to %q.\n", args[0])
				return nil
			}

			// No args, no --clear: show current context.
			ctx, err := properties.Load(opts.ConfigDir)
			if err != nil {
				return fmt.Errorf("failed to load context: %w", err)
			}

			if ctx.App == "" {
				output.Writef(cmd.OutOrStdout(), "No active app context. Use 'admiral use <app-name>' to set one.\n")
				return nil
			}

			output.Writef(cmd.OutOrStdout(), "%s\n", ctx.App)
			return nil
		},
	}

	cmd.Flags().BoolVar(&clear, "clear", false, "clear the active app context")

	return cmd
}
