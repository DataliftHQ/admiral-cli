package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newTokenGetCmd(opts *factory.Options) *cobra.Command {
	var (
		clusterID string
		tokenID   string
	)

	cmd := &cobra.Command{
		Use:   "get [cluster] [token]",
		Short: "Get a cluster token",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clusterName := ""
			if len(args) > 0 {
				clusterName = args[0]
			}
			if clusterName == "" && clusterID == "" {
				return fmt.Errorf("provide a cluster name or use --cluster-id")
			}

			tokenName := ""
			if len(args) > 1 {
				tokenName = args[1]
			}
			if tokenName == "" && tokenID == "" {
				return fmt.Errorf("provide a token name as the second argument or use --id")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resolvedClusterID, err := cmdutil.ResolveClusterID(cmd.Context(), c.Cluster(), clusterName, clusterID)
			if err != nil {
				return err
			}

			resolvedTokenID, err := cmdutil.ResolveClusterTokenID(cmd.Context(), c.Cluster(), resolvedClusterID, tokenName, tokenID)
			if err != nil {
				return err
			}

			resp, err := c.Cluster().GetClusterToken(cmd.Context(), &clusterv1.GetClusterTokenRequest{
				ClusterId: resolvedClusterID,
				TokenId:   resolvedTokenID,
			})
			if err != nil {
				return err
			}

			t := resp.AccessToken
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "ID", Value: t.Id},
						{Key: "Name", Value: t.Name},
						{Key: "Token Type", Value: output.FormatEnum(t.TokenType.String(), "TOKEN_TYPE_")},
						{Key: "Status", Value: output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_")},
						{Key: "Scopes", Value: output.FormatScopes(t.Scopes)},
						{Key: "Binding Type", Value: output.FormatEnum(t.BindingType.String(), "BINDING_TYPE_")},
						{Key: "Binding ID", Value: t.BindingId},
						{Key: "Created", Value: output.FormatTimestamp(t.CreatedAt)},
						{Key: "Created By", Value: t.CreatedBy},
						{Key: "Expires", Value: output.FormatTimestamp(t.ExpiresAt)},
						{Key: "Last Used", Value: output.FormatTimestamp(t.LastUsedAt)},
						{Key: "Revoked", Value: output.FormatTimestamp(t.RevokedAt)},
						{Key: "Age", Value: output.FormatAge(t.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}

	cmd.Flags().StringVar(&clusterID, "cluster-id", "", "cluster UUID (bypasses name resolution)")
	cmd.Flags().StringVar(&tokenID, "id", "", "token UUID (bypasses name resolution)")

	return cmd
}
