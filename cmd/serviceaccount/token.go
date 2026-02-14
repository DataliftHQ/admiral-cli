package serviceaccount

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

func newTokenCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "token",
		Short:         "Manage service account tokens",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newTokenCreateCmd(opts),
		newTokenListCmd(opts),
		newTokenGetCmd(opts),
		newTokenRevokeCmd(opts),
	)

	return cmd
}
