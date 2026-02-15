package auth

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// AuthCmd is the parent command for authentication operations.
type AuthCmd struct {
	Cmd *cobra.Command
}

// NewAuthCmd creates the auth command tree.
func NewAuthCmd(opts *factory.Options) *AuthCmd {
	root := &AuthCmd{}

	cmd := &cobra.Command{
		Use:           "auth",
		Short:         "Manage authentication",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		NewLoginCmd(opts),
		NewLogoutCmd(opts),
		newStatusCmd(opts),
		newWhoamiCmd(opts),
	)

	root.Cmd = cmd
	return root
}
