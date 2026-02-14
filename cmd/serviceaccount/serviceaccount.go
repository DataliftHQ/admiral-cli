package serviceaccount

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// ServiceAccountCmd is the parent command for service account operations.
type ServiceAccountCmd struct {
	Cmd *cobra.Command
}

// NewServiceAccountCmd creates the service-account command tree.
func NewServiceAccountCmd(opts *factory.Options) *ServiceAccountCmd {
	root := &ServiceAccountCmd{}

	cmd := &cobra.Command{
		Use:           "service-account",
		Short:         "Manage service accounts",
		Aliases:       []string{"sa"},
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	// Token sub-group
	tokenCmd := newTokenCmd(opts)

	cmd.AddCommand(
		newListCmd(opts),
		newGetCmd(opts),
		newCreateCmd(opts),
		newUpdateCmd(opts),
		newDeleteCmd(opts),
		tokenCmd,
	)

	root.Cmd = cmd
	return root
}
