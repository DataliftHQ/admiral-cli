package cluster

import (
	"fmt"
	"text/tabwriter"

	commonv1 "buf.build/gen/go/admiral/common/protocolbuffers/go/admiral/common/v1"
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	clusterv1 "go.admiral.io/sdk/proto/admiral/api/cluster/v1"
)

func newTokenRevokeCmd(opts *factory.Options) *cobra.Command {
	var (
		clusterID string
		tokenID   string
		confirm   bool
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "revoke [cluster] [token]",
		Short: "Revoke a cluster token",
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

			displayName := tokenName
			if displayName == "" {
				displayName = tokenID
			}

			if !confirm {
				return fmt.Errorf("use --confirm to revoke token %s", displayName)
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

			if !force {
				if err := checkLastActiveToken(cmd, c.Cluster(), resolvedClusterID, resolvedTokenID); err != nil {
					return err
				}
			}

			resp, err := c.Cluster().RevokeClusterToken(cmd.Context(), &clusterv1.RevokeClusterTokenRequest{
				ClusterId: resolvedClusterID,
				TokenId:   resolvedTokenID,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				output.Writef(w, "Token %s revoked\n", displayName)
			})
		},
	}

	cmd.Flags().StringVar(&clusterID, "cluster-id", "", "cluster UUID (bypasses name resolution)")
	cmd.Flags().StringVar(&tokenID, "id", "", "token UUID (bypasses name resolution)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm token revocation")
	cmd.Flags().BoolVar(&force, "force", false, "skip last-token safety check")

	return cmd
}

// checkLastActiveToken returns an error if tokenID is the only active token
// for the cluster, preventing accidental loss of agent connectivity.
func checkLastActiveToken(cmd *cobra.Command, client clusterv1.ClusterAPIClient, clusterID, tokenID string) error {
	resp, err := client.ListClusterTokens(cmd.Context(), &clusterv1.ListClusterTokensRequest{
		ClusterId: clusterID,
		Filter:    "field['status'] = 'ACTIVE'",
	})
	if err != nil {
		return fmt.Errorf("checking active tokens: %w", err)
	}

	// Client-side filter: the API may not apply the status filter.
	var activeIDs []string
	for _, t := range resp.AccessTokens {
		if t.Status == commonv1.AccessTokenStatus_ACCESS_TOKEN_STATUS_ACTIVE {
			activeIDs = append(activeIDs, t.Id)
		}
	}

	if len(activeIDs) == 1 && activeIDs[0] == tokenID {
		return fmt.Errorf(
			"this is the last active token for the cluster; revoking it will disconnect the agent\n\nUse --force to proceed anyway",
		)
	}

	if len(activeIDs) <= 2 {
		output.Writef(cmd.ErrOrStderr(),
			"Warning: only %d active token(s) will remain after this revocation\n", len(activeIDs)-1)
	}

	return nil
}
