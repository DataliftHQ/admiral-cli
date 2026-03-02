package env

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var (
		appFlag string
		envID   string
		confirm bool
	)

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an environment",
		Long: `Delete an environment.

The environment is identified by name (positional argument) within the
parent application. Use --id to look up by UUID directly.

Requires --confirm to prevent accidental deletion.`,
		Example: `  # Delete an environment
  admiral env delete staging --app billing-api --confirm

  # Delete by UUID
  admiral env delete --id 550e8400-e29b-41d4-a716-446655440000 --confirm`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName string
			if len(args) == 1 {
				envName = args[0]
			}
			if envName == "" && envID == "" {
				return fmt.Errorf("environment name or --id is required")
			}

			display := envName
			if display == "" {
				display = envID
			}

			if !confirm {
				return fmt.Errorf("use --confirm to delete environment %s", display)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			id := envID
			if id == "" {
				appName, err := resolveAppForEnv(appFlag)
				if err != nil {
					return err
				}
				appID, err := cmdutil.ResolveAppID(cmd.Context(), c.Application(), appName, "")
				if err != nil {
					return err
				}
				id, err = cmdutil.ResolveEnvID(cmd.Context(), c.Environment(), appID, envName, "")
				if err != nil {
					return err
				}
			}

			resp, err := c.Environment().DeleteEnvironment(cmd.Context(), &environmentv1.DeleteEnvironmentRequest{
				EnvironmentId: id,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Environment %s deleted\n", display)
			})
		},
	}

	addAppFlag(cmd, &appFlag)
	cmd.Flags().StringVar(&envID, "id", "", "environment ID (UUID)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")

	return cmd
}
