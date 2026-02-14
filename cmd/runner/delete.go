package runner

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <runner-id>",
		Short: "Delete a runner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("use --confirm to delete the runner %q", args[0])
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			_, err = c.Runner().DeleteRunner(cmd.Context(), &runnerv1.DeleteRunnerRequest{
				RunnerId: args[0],
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Runner %s deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")

	return cmd
}
