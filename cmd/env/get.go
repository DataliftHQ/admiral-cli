package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	var (
		appFlag string
		envID   string
	)

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get environment details",
		Long: `Get detailed information about an environment.

The environment is identified by name (positional argument) within the
parent application. Use --id to look up by UUID directly.`,
		Example: `  # Get environment details
  admiral env get staging --app billing-api

  # Get environment by UUID
  admiral env get --id 550e8400-e29b-41d4-a716-446655440000`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName string
			if len(args) == 1 {
				envName = args[0]
			}
			if envName == "" && envID == "" {
				return fmt.Errorf("environment name or --id is required")
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

			resp, err := c.Environment().GetEnvironment(cmd.Context(), &environmentv1.GetEnvironmentRequest{
				EnvironmentId: id,
			})
			if err != nil {
				return err
			}

			env := resp.Environment
			p := output.NewPrinter(opts.OutputFormat)

			details := []output.Detail{
				{Key: "ID", Value: env.Id},
				{Key: "Name", Value: env.Name},
				{Key: "Application", Value: env.ApplicationId},
				{Key: "Runtime Type", Value: output.FormatEnum(env.RuntimeType.String(), "RUNTIME_TYPE_")},
			}

			if k := env.GetKubernetes(); k != nil {
				details = append(details,
					output.Detail{Key: "Cluster ID", Value: k.ClusterId},
					output.Detail{Key: "Namespace", Value: k.GetNamespace()},
				)
			}

			if infra := env.GetInfrastructure(); infra != nil {
				details = append(details, output.Detail{Key: "Runner", Value: infra.RunnerId})
			}

			details = append(details,
				output.Detail{Key: "Description", Value: env.Description},
				output.Detail{Key: "Labels", Value: output.FormatLabels(env.Labels)},
				output.Detail{Key: "Has Pending Changes", Value: fmt.Sprintf("%t", env.HasPendingChanges)},
				output.Detail{Key: "Last Deployed", Value: output.FormatTimestamp(env.LastDeployedAt)},
				output.Detail{Key: "Created", Value: output.FormatTimestamp(env.CreatedAt)},
				output.Detail{Key: "Created By", Value: env.CreatedBy},
				output.Detail{Key: "Updated", Value: output.FormatTimestamp(env.UpdatedAt)},
				output.Detail{Key: "Updated By", Value: env.UpdatedBy},
				output.Detail{Key: "Age", Value: output.FormatAge(env.CreatedAt)},
			)

			sections := []output.Section{
				{Details: details},
			}

			return p.PrintDetail(resp, sections)
		},
	}

	addAppFlag(cmd, &appFlag)
	cmd.Flags().StringVar(&envID, "id", "", "environment ID (UUID)")

	return cmd
}
