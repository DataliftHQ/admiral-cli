package whoami

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/credentials"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
)

// userResponse is the JSON shape returned by the REST API.
type userResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	GivenName   string `json:"givenName"`
	FamilyName  string `json:"familyName"`
	AvatarURL   string `json:"avatarUrl"`
}

// NewWhoamiCmd creates the whoami command.
func NewWhoamiCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current user information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := credentials.ResolveToken(opts.ConfigDir)
			if err != nil {
				return err
			}

			user, err := fetchUser(token)
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Email", Value: user.Email},
						{Key: "Display Name", Value: user.DisplayName},
						{Key: "Given Name", Value: user.GivenName},
						{Key: "Family Name", Value: user.FamilyName},
						{Key: "ID", Value: user.ID},
						{Key: "Tenant", Value: user.TenantID},
					},
				},
			}

			return p.PrintDetail(nil, sections)
		},
	}
}

func fetchUser(token string) (*userResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://app.admiral.io/api/v1/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach API: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort cleanup

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	var user userResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &user, nil
}
