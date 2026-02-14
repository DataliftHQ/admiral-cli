package runner

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	accesstokenv1 "go.admiral.io/sdk/proto/accesstoken/v1"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newTokenGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <runner-id> <token-id>",
		Short: "Get a runner access token",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Runner().GetRunnerToken(cmd.Context(), &runnerv1.GetRunnerTokenRequest{
				RunnerId: args[0],
				TokenId:  args[1],
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				printTokenTable(w, []*accesstokenv1.AccessToken{resp.AccessToken})
			})
		},
	}

	return cmd
}
