package token

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/user/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		name      string
		scopes    []string
		expiresAt string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a personal access token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &userv1.CreatePersonalAccessTokenRequest{
				DisplayName: name,
				Scopes:      scopes,
			}

			if expiresAt != "" {
				t, err := time.Parse(time.RFC3339, expiresAt)
				if err != nil {
					return fmt.Errorf("invalid --expires-at format: %w (expected RFC3339, e.g. 2025-12-31T23:59:59Z)", err)
				}
				req.ExpiresAt = timestamppb.New(t)
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.User().CreatePersonalAccessToken(cmd.Context(), req)
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

	cmd.Flags().StringVar(&name, "name", "", "display name for the token (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "comma-separated list of scopes")
	cmd.Flags().StringVar(&expiresAt, "expires-at", "", "expiration date in RFC3339 format (e.g. 2025-12-31T23:59:59Z)")

	return cmd
}
