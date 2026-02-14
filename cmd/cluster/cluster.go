package cluster

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

// ClusterCmd is the parent command for cluster operations.
type ClusterCmd struct {
	Cmd *cobra.Command
}

// NewClusterCmd creates the cluster command tree.
func NewClusterCmd(opts *factory.Options) *ClusterCmd {
	root := &ClusterCmd{}

	cmd := &cobra.Command{
		Use:           "cluster",
		Short:         "Manage clusters",
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
		newStatusCmd(opts),
		tokenCmd,
	)

	root.Cmd = cmd
	return root
}
