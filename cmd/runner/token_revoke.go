package runner

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newTokenRevokeCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "revoke <runner-id> <token-id>",
		Short: "Revoke a runner access token",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("use --confirm to revoke token %q for runner %q", args[1], args[0])
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			_, err = c.Runner().RevokeRunnerToken(cmd.Context(), &runnerv1.RevokeRunnerTokenRequest{
				RunnerId: args[0],
				TokenId:  args[1],
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Token %s revoked for runner %s\n", args[1], args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm revocation")

	return cmd
}
