package serviceaccount

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newDeleteCmd(opts *factory.Options) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <service-account-id>",
		Short: "Delete a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !confirm {
				return fmt.Errorf("use --confirm to delete service account %s", id)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().DeleteServiceAccount(cmd.Context(), &serviceaccountv1.DeleteServiceAccountRequest{
				ServiceAccountId: id,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				_, _ = fmt.Fprintf(w, "Service account %s deleted\n", id)
			})
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "confirm service account deletion")

	return cmd
}
