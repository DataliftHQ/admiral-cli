package runner

import (
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	accesstokenv1 "go.admiral.io/sdk/proto/accesstoken/v1"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newTokenCreateCmd(opts *factory.Options) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create <runner-id>",
		Short: "Create a runner access token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Runner().CreateRunnerToken(cmd.Context(), &runnerv1.CreateRunnerTokenRequest{
				RunnerId:    args[0],
				DisplayName: name,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				printTokenTable(w, []*accesstokenv1.AccessToken{resp.AccessToken})
			}); err != nil {
				return err
			}

			if resp.PlainTextToken != "" {
				output.PrintToken(cmd.ErrOrStderr(), resp.PlainTextToken)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the token")

	return cmd
}
