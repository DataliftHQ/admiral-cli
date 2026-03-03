package cluster

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
)

func newWorkloadCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "workload",
		Short:         "Manage cluster workloads",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
	}

	cmd.AddCommand(
		newWorkloadListCmd(opts),
	)

	return cmd
}
