package token

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	userv1 "go.admiral.io/sdk/proto/admiral/api/user/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		scopes    []string
		expiresIn string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a personal access token",
		Args:  cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := &userv1.CreatePersonalAccessTokenRequest{
				Name:   args[0],
				Scopes: scopes,
			}

			if expiresIn != "" {
				d, err := parseDuration(expiresIn)
				if err != nil {
					return fmt.Errorf("invalid --expires-in value: %w", err)
				}
				req.ExpiresAt = timestamppb.New(time.Now().Add(d))
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
				output.Writeln(w, "ID\tNAME\tSTATUS\tCREATED")
				output.Writef(w, "%s\t%s\t%s\t%s\n",
					t.Id,
					t.Name,
					output.FormatEnum(t.Status.String(), "ACCESS_TOKEN_STATUS_"),
					output.FormatTimestamp(t.CreatedAt),
				)
			}); err != nil {
				return err
			}

			output.PrintToken(cmd.ErrOrStderr(), resp.PlainTextToken)

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&scopes, "scope", nil, "token scope (repeatable)")
	cmd.Flags().StringVar(&expiresIn, "expires-in", "", "token lifetime (e.g. 24h, 30d, 90d, 1y)")

	return cmd
}

// parseDuration parses a duration string supporting Go durations (24h, 720h)
// plus day (30d) and year (1y) suffixes.
func parseDuration(s string) (time.Duration, error) {
	if v, ok := strings.CutSuffix(s, "d"); ok {
		days, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid day value: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	if v, ok := strings.CutSuffix(s, "y"); ok {
		years, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid year value: %s", s)
		}
		return time.Duration(years) * 365 * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
