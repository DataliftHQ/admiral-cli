package serviceaccount

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newTokenCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		name   string
		scopes []string
	)

	cmd := &cobra.Command{
		Use:   "create <service-account-id>",
		Short: "Create a token for a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			saID := args[0]

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().CreateServiceAccountToken(cmd.Context(), &serviceaccountv1.CreateServiceAccountTokenRequest{
				ServiceAccountId: saID,
				DisplayName:      name,
				Scopes:           scopes,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				t := resp.AccessToken
				_, _ = fmt.Fprintln(w, "ID\tNAME\tPREFIX\tSTATUS\tCREATED")
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					t.Id,
					t.DisplayName,
					t.TokenPrefix,
					output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
					output.FormatTimestamp(t.CreatedAt),
				)
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
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "comma-separated list of scopes")

	return cmd
}
