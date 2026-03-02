package env

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		appFlag     string
		envID       string
		newName     string
		description string
		runtimeType string
		cluster     string
		namespace   string
		runner      string
		labelStrs   []string
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an environment",
		Long: `Update an existing environment.

The environment is identified by name (positional argument) within the
parent application. Use --id to look up by UUID directly.`,
		Example: `  # Update description
  admiral env update staging --app billing-api --description "New description"

  # Update labels
  admiral env update staging --app billing-api --label tier=production

  # Update cluster binding
  admiral env update staging --app billing-api --cluster new-cluster

  # Update by UUID
  admiral env update --id 550e8400-e29b-41d4-a716-446655440000 --description "Updated"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName string
			if len(args) == 1 {
				envName = args[0]
			}
			if envName == "" && envID == "" {
				return fmt.Errorf("environment name or --id is required")
			}

			var paths []string
			environment := &environmentv1.Environment{}

			if cmd.Flags().Changed("name") {
				environment.Name = newName
				paths = append(paths, "name")
			}

			if cmd.Flags().Changed("description") {
				environment.Description = description
				paths = append(paths, "description")
			}

			if cmd.Flags().Changed("runtime-type") {
				rt, err := parseRuntimeType(runtimeType)
				if err != nil {
					return err
				}
				environment.RuntimeType = rt
				paths = append(paths, "runtime_type")
			}

			if cmd.Flags().Changed("label") {
				labels, err := cmdutil.ParseLabels(labelStrs)
				if err != nil {
					return err
				}
				environment.Labels = labels
				paths = append(paths, "labels")
			}

			if cmd.Flags().Changed("runner") {
				environment.Infrastructure = &environmentv1.InfrastructureConfig{
					RunnerId: runner,
				}
				paths = append(paths, "infrastructure")
			}

			if cmd.Flags().Changed("namespace") && !cmd.Flags().Changed("cluster") {
				return fmt.Errorf("--namespace requires --cluster")
			}

			if len(paths) == 0 && !cmd.Flags().Changed("cluster") && !cmd.Flags().Changed("namespace") {
				return fmt.Errorf("at least one field must be specified for update")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			if cmd.Flags().Changed("cluster") || cmd.Flags().Changed("namespace") {
				k8sCfg := &environmentv1.KubernetesConfig{}

				if cmd.Flags().Changed("cluster") {
					clusterID, err := cmdutil.ResolveClusterID(cmd.Context(), c.Cluster(), cluster)
					if err != nil {
						return err
					}
					k8sCfg.ClusterId = clusterID
				}

				if cmd.Flags().Changed("namespace") {
					k8sCfg.Namespace = &namespace
				}

				environment.RuntimeConfig = &environmentv1.Environment_Kubernetes{
					Kubernetes: k8sCfg,
				}
				paths = append(paths, "runtime_config")
			}

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

			environment.Id = id

			resp, err := c.Environment().UpdateEnvironment(cmd.Context(), &environmentv1.UpdateEnvironmentRequest{
				Environment: environment,
				UpdateMask:  &fieldmaskpb.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				env := resp.Environment
				output.Writeln(w, "NAME\tRUNTIME\tDESCRIPTION\tAGE")
				output.Writef(w, "%s\t%s\t%s\t%s\n",
					env.Name,
					output.FormatEnum(env.RuntimeType.String(), "RUNTIME_TYPE_"),
					env.Description,
					output.FormatAge(env.CreatedAt),
				)
			})
		},
	}

	addAppFlag(cmd, &appFlag)
	cmd.Flags().StringVar(&envID, "id", "", "environment ID (UUID)")
	cmd.Flags().StringVar(&newName, "name", "", "new environment name")
	cmd.Flags().StringVar(&description, "description", "", "environment description")
	cmd.Flags().StringVar(&runtimeType, "runtime-type", "", "runtime type (kubernetes)")
	cmd.Flags().StringVar(&cluster, "cluster", "", "target cluster name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace")
	cmd.Flags().StringVar(&runner, "runner", "", "runner ID for infrastructure operations")
	cmdutil.AddLabelFlag(cmd, &labelStrs, "label to set (key=value, repeatable)")

	return cmd
}
