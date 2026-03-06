package token

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/admiral/api/user/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	var tokenID string

	cmd := &cobra.Command{
		Use:   "get [token]",
		Short: "Get a personal access token",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenName := ""
			if len(args) > 0 {
				tokenName = args[0]
			}
			if tokenName == "" && tokenID == "" {
				return fmt.Errorf("provide a token name or use --id")
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resolvedID, err := cmdutil.ResolvePersonalAccessTokenID(cmd.Context(), c.User(), tokenName, tokenID)
			if err != nil {
				return err
			}

			resp, err := c.User().GetPersonalAccessToken(cmd.Context(), &userv1.GetPersonalAccessTokenRequest{
				TokenId: resolvedID,
			})
			if err != nil {
				return err
			}

			t := resp.AccessToken
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "ID", Value: t.Id},
						{Key: "Name", Value: t.Name},
						{Key: "Token Type", Value: output.FormatEnum(t.TokenType.String(), "TOKEN_TYPE_")},
						{Key: "Status", Value: output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_")},
						{Key: "Scopes", Value: output.FormatScopes(t.Scopes)},
						{Key: "Created", Value: output.FormatTimestamp(t.CreatedAt)},
						{Key: "Created By", Value: t.CreatedBy},
						{Key: "Expires", Value: output.FormatTimestamp(t.ExpiresAt)},
						{Key: "Last Used", Value: output.FormatTimestamp(t.LastUsedAt)},
						{Key: "Revoked", Value: output.FormatTimestamp(t.RevokedAt)},
						{Key: "Age", Value: output.FormatAge(t.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}

	cmd.Flags().StringVar(&tokenID, "id", "", "token UUID (bypasses name resolution)")

	return cmd
}
