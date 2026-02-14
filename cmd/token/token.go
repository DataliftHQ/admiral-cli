package token

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// TokenCmd is the parent command for personal access token operations.
type TokenCmd struct {
	Cmd *cobra.Command
}

// NewTokenCmd creates the token command tree.
func NewTokenCmd(opts *factory.Options) *TokenCmd {
	root := &TokenCmd{}

	cmd := &cobra.Command{
		Use:           "token",
		Short:         "Manage personal access tokens",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newListCmd(opts),
		newGetCmd(opts),
		newCreateCmd(opts),
		newRevokeCmd(opts),
	)

	root.Cmd = cmd
	return root
}
