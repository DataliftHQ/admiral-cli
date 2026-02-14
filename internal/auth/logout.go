package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.admiral.io/cli/internal/credentials"
)

// LogoutOptions configures the logout flow.
type LogoutOptions struct {
	Issuer    string
	ClientID  string
	ConfigDir string
}

// Logout revokes the refresh token and deletes stored credentials.
func Logout(_ context.Context, opts LogoutOptions) error {
	// Get current token before deleting (for revocation).
	token, _ := credentials.GetToken(opts.ConfigDir)

	// Delete stored credentials.
	if err := credentials.DeleteToken(opts.ConfigDir); err != nil {
		return err
	}

	// Attempt to revoke the refresh token (best-effort). The access token
	// is short-lived and will expire on its own.
	if token != nil && token.RefreshToken != "" {
		_ = revokeRefreshToken(token.RefreshToken, opts.ClientID, opts.Issuer)
	}

	return nil
}

func revokeRefreshToken(token, clientID, issuer string) error {
	u, err := url.Parse(issuer)
	if err != nil {
		return err
	}
	u.Path += "/oauth2/revoke"

	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", clientID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(u.String(), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort cleanup

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token revocation failed (HTTP %d)", resp.StatusCode)
	}

	return nil
}
