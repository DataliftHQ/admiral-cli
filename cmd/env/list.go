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

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		appFlag   string
		pageSize  int32
		pageToken string
		labelStrs []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List environments",
		Long:  `List all environments for an application.`,
		Example: `  # List environments
  admiral env list --app billing-api

  # List with label filter
  admiral env list --app billing-api --label region=us-east-1

  # Paginated listing
  admiral env list --app billing-api --page-size 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			appName, err := resolveAppForEnv(appFlag)
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

			// Build filter: always scope to application, optionally add labels.
			filters := []string{fmt.Sprintf("field['application_id'] = '%s'", appID)}
			if len(labelStrs) > 0 {
				labelFilter, err := cmdutil.BuildLabelFilter(labelStrs)
				if err != nil {
					return err
				}
				filters = append(filters, labelFilter)
			}

			resp, err := c.Environment().ListEnvironments(cmd.Context(), &environmentv1.ListEnvironmentsRequest{
				PageSize:  pageSize,
				PageToken: pageToken,
				Filter:    strings.Join(filters, " AND "),
			})
			if err != nil {
				return err
			}

			if len(resp.Environments) == 0 {
				output.Writef(cmd.OutOrStdout(), "No environments found for application %s\n", appName)
				return nil
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				if opts.OutputFormat == output.FormatWide {
					output.Writeln(w, "APP ID\tENV ID\tNAME\tDESCRIPTION\tRUNTIME\tRUNTIME CONFIG\tLABELS\tAGE")
					for _, env := range resp.Environments {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
							env.ApplicationId,
							env.Id,
							env.Name,
							env.Description,
							output.FormatEnum(env.RuntimeType.String(), "RUNTIME_TYPE_"),
							formatRuntimeConfig(env),
							output.FormatLabels(env.Labels),
							output.FormatAge(env.CreatedAt),
						)
					}
				} else {
					output.Writeln(w, "NAME\tDESCRIPTION\tRUNTIME\tLABELS\tAGE")
					for _, env := range resp.Environments {
						output.Writef(w, "%s\t%s\t%s\t%s\t%s\n",
							env.Name,
							env.Description,
							output.FormatEnum(env.RuntimeType.String(), "RUNTIME_TYPE_"),
							output.FormatLabels(env.Labels),
							output.FormatAge(env.CreatedAt),
						)
					}
				}
			}); err != nil {
				return err
			}

			if resp.NextPageToken != "" && opts.OutputFormat != output.FormatJSON && opts.OutputFormat != output.FormatYAML {
				output.Writef(cmd.ErrOrStderr(), "\nNEXT PAGE TOKEN: %s\n", resp.NextPageToken)
			}

			return nil
		},
	}

	addAppFlag(cmd, &appFlag)
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "maximum number of results per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "pagination token from a previous response")
	cmdutil.AddLabelFlag(cmd, &labelStrs, "filter by label (key=value, repeatable)")

	return cmd
}

// formatRuntimeConfig returns a compact string representation of the
// runtime-specific config. Adapts automatically to the runtime type.
func formatRuntimeConfig(env *environmentv1.Environment) string {
	if k := env.GetKubernetes(); k != nil {
		var parts []string
		if k.ClusterId != "" {
			parts = append(parts, "cluster="+k.ClusterId)
		}
		if ns := k.GetNamespace(); ns != "" {
			parts = append(parts, "namespace="+ns)
		}
		if len(parts) > 0 {
			return strings.Join(parts, ", ")
		}
	}
	return ""
}
