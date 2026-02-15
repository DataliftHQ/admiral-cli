package cluster

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/cluster/v1"
)

func newTokenCreateCmd(opts *factory.Options) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create <cluster-id>",
		Short: "Create a cluster token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterID := args[0]

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Cluster().CreateClusterToken(cmd.Context(), &clusterv1.CreateClusterTokenRequest{
				ClusterId:   clusterID,
				DisplayName: name,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				t := resp.AccessToken
				output.Writeln(w, "ID\tNAME\tPREFIX\tSTATUS\tCREATED")
				output.Writef(w, "%s\t%s\t%s\t%s\t%s\n",
					t.Id,
					t.DisplayName,
					t.TokenPrefix,
					output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
					output.FormatTimestamp(t.CreatedAt),
				)
			}); err != nil {
				return err
			}

			output.PrintToken(cmd.ErrOrStderr(), resp.PlainTextToken)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the token (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
