package serviceaccount

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	serviceaccountv1 "go.admiral.io/sdk/proto/serviceaccount/v1"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		name        string
		description string
		status      string
	)

	cmd := &cobra.Command{
		Use:   "update <service-account-id>",
		Short: "Update a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			var paths []string
			sa := &serviceaccountv1.ServiceAccount{Id: id}

			if cmd.Flags().Changed("name") {
				sa.DisplayName = name
				paths = append(paths, "display_name")
			}

			if cmd.Flags().Changed("description") {
				sa.Description = description
				paths = append(paths, "description")
			}

			if cmd.Flags().Changed("status") {
				switch strings.ToLower(status) {
				case "active":
					sa.Status = serviceaccountv1.ServiceAccountStatus_SERVICE_ACCOUNT_STATUS_ACTIVE
				case "disabled":
					sa.Status = serviceaccountv1.ServiceAccountStatus_SERVICE_ACCOUNT_STATUS_DISABLED
				default:
					return fmt.Errorf("invalid status %q: must be one of active, disabled", status)
				}
				paths = append(paths, "status")
			}

			if len(paths) == 0 {
				return fmt.Errorf("at least one of --name, --description, or --status must be specified")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.ServiceAccount().UpdateServiceAccount(cmd.Context(), &serviceaccountv1.UpdateServiceAccountRequest{
				ServiceAccount: sa,
				UpdateMask:     &fieldmaskpb.FieldMask{Paths: paths},
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			return p.PrintResource(resp, func(w *tabwriter.Writer) {
				updated := resp.ServiceAccount
				if opts.OutputFormat == output.FormatWide {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tAGE\tDESCRIPTION\tSCOPES\tCREATED\tUPDATED")
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						updated.Id,
						updated.DisplayName,
						output.FormatEnum(updated.Status.String(), "SERVICE_ACCOUNT_STATUS_"),
						output.FormatAge(updated.CreatedAt),
						updated.Description,
						strings.Join(updated.Scopes, ","),
						output.FormatTimestamp(updated.CreatedAt),
						output.FormatTimestamp(updated.UpdatedAt),
					)
				} else {
					_, _ = fmt.Fprintln(w, "ID\tNAME\tSTATUS\tAGE")
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						updated.Id,
						updated.DisplayName,
						output.FormatEnum(updated.Status.String(), "SERVICE_ACCOUNT_STATUS_"),
						output.FormatAge(updated.CreatedAt),
					)
				}
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new display name for the service account")
	cmd.Flags().StringVar(&description, "description", "", "new description for the service account")
	cmd.Flags().StringVar(&status, "status", "", "new status: active or disabled")

	return cmd
}
