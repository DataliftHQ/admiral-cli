package env

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	environmentv1 "go.admiral.io/sdk/proto/admiral/api/environment/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		appFlag     string
		description string
		runtimeType string
		cluster     string
		namespace   string
		runner      string
		labelStrs   []string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an environment",
		Long:  `Create a new environment within an application.`,
		Example: `  # Create an environment with Kubernetes runtime
  admiral env create staging --app billing-api --runtime-type kubernetes --cluster prod-cluster

  # Create with namespace and labels
  admiral env create staging --app billing-api --runtime-type kubernetes --cluster prod-cluster --namespace staging-ns --label region=us-east-1

  # Create with description
  admiral env create staging --app billing-api --description "Staging environment"`,
		Args: cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			appName, err := resolveAppForEnv(appFlag)
			if err != nil {
				return err
			}

			labels, err := cmdutil.ParseLabels(labelStrs)
			if err != nil {
				return err
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			appID, err := cmdutil.ResolveAppID(cmd.Context(), c.Application(), appName, "")
			if err != nil {
				return err
			}

			req := &environmentv1.CreateEnvironmentRequest{
				ApplicationId: appID,
				Name:          name,
				Labels:        labels,
			}

			if cmd.Flags().Changed("description") {
				req.Description = description
			}

			if cmd.Flags().Changed("runtime-type") {
				rt, err := parseRuntimeType(runtimeType)
				if err != nil {
					return err
				}
				req.RuntimeType = rt
			}

			if cmd.Flags().Changed("namespace") && !cmd.Flags().Changed("cluster") {
				return fmt.Errorf("--namespace requires --cluster")
			}

			if cmd.Flags().Changed("cluster") || cmd.Flags().Changed("namespace") {
				k8sCfg := &environmentv1.KubernetesConfig{}

				if cmd.Flags().Changed("cluster") {
					clusterID, err := cmdutil.ResolveClusterID(cmd.Context(), c.Cluster(), cluster, "")
					if err != nil {
						return err
					}
					k8sCfg.ClusterId = clusterID
				}

				if cmd.Flags().Changed("namespace") {
					k8sCfg.Namespace = &namespace
				}

				req.RuntimeConfig = &environmentv1.CreateEnvironmentRequest_Kubernetes{
					Kubernetes: k8sCfg,
				}
			}

			if cmd.Flags().Changed("runner") {
				req.Infrastructure = &environmentv1.InfrastructureConfig{
					RunnerId: runner,
				}
			}

			resp, err := c.Environment().CreateEnvironment(cmd.Context(), req)
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
	cmd.Flags().StringVar(&description, "description", "", "environment description")
	cmd.Flags().StringVar(&runtimeType, "runtime-type", "", "runtime type (kubernetes)")
	cmd.Flags().StringVar(&cluster, "cluster", "", "target cluster name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace")
	cmd.Flags().StringVar(&runner, "runner", "", "runner ID for infrastructure operations")
	cmdutil.AddLabelFlag(cmd, &labelStrs, "label to attach (key=value, repeatable)")

	return cmd
}

// parseRuntimeType converts a CLI string to the proto RuntimeType enum.
func parseRuntimeType(s string) (environmentv1.RuntimeType, error) {
	switch strings.ToLower(s) {
	case "kubernetes", "k8s":
		return environmentv1.RuntimeType_RUNTIME_TYPE_KUBERNETES, nil
	default:
		return environmentv1.RuntimeType_RUNTIME_TYPE_UNSPECIFIED,
			fmt.Errorf("unsupported runtime type %q; valid values: kubernetes", s)
	}
}
