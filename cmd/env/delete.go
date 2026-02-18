package env

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an environment",
		Long: `Delete an environment.

Requires --confirm to prevent accidental deletion.`,
		Example: `  # Delete an environment
  admiral use billing-api
  admiral env delete staging --confirm`,
		Args: cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			app, err := resolveAppForEnv(opts.ConfigDir)
			if err != nil {
				return err
			}

			if !confirm {
				return fmt.Errorf("--confirm is required to delete an environment")
			}

			stub := cmdutil.StubResult{
				Command:     "env delete",
				App:         app,
				Environment: slug,
				Status:      cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")

	return cmd
}
