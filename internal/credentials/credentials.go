package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"

	"go.admiral.io/sdk/client"
)

const (
	// EnvToken is the environment variable for providing an access token directly.
	EnvToken = "ADMIRAL_TOKEN"

	// credentialsFile is the name of the file where tokens are stored.
	credentialsFile = "credentials.json"

	// refreshWindow is how far before expiry we proactively refresh.
	refreshWindow = 30 * time.Second
)

// credentials holds an OAuth2 token and the metadata needed for refresh.
type credentials struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`

	// Stored so ResolveToken can refresh without external config.
	ClientID string `json:"client_id,omitempty"`
	TokenURL string `json:"token_url,omitempty"`
}

// TokenResult holds a resolved token and its auth scheme.
type TokenResult struct {
	Token      string
	AuthScheme client.AuthScheme
}

// ResolveToken returns a valid access token from the environment variable or
// credentials file. If the stored token is expired (or close to expiry) and a
// refresh token is available, it transparently refreshes and persists the new token.
func ResolveToken(configDir string) (*TokenResult, error) {
	if t := os.Getenv(EnvToken); t != "" {
		return &TokenResult{Token: t, AuthScheme: client.AuthSchemeToken}, nil
	}

	creds, err := readCredentials(configDir)
	if err != nil {
		return nil, fmt.Errorf("not logged in: run 'admiral auth login' first")
	}

	if creds.AccessToken == "" {
		return nil, fmt.Errorf("not logged in: run 'admiral auth login' first")
	}

	// If the token is still valid, return it directly.
	if creds.Expiry.IsZero() || time.Until(creds.Expiry) > refreshWindow {
		return &TokenResult{Token: creds.AccessToken, AuthScheme: client.AuthSchemeBearer}, nil
	}

	// Token is expired or about to expire — try to refresh.
	if creds.RefreshToken == "" || creds.TokenURL == "" {
		return nil, fmt.Errorf("session expired: run 'admiral auth login' to re-authenticate")
	}

	refreshed, err := refreshCredentials(configDir, creds)
	if err != nil {
		return nil, fmt.Errorf("session expired (refresh failed): run 'admiral auth login' to re-authenticate")
	}

	return &TokenResult{Token: refreshed.AccessToken, AuthScheme: client.AuthSchemeBearer}, nil
}

// SaveToken writes an OAuth2 token and refresh metadata to the credentials file.
func SaveToken(configDir string, token *oauth2.Token, clientID, tokenURL string) error {
	creds := &credentials{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ClientID:     clientID,
		TokenURL:     tokenURL,
	}

	return writeCredentials(configDir, creds)
}

// GetToken reads the OAuth2 token from the credentials file.
// Returns nil if no credentials are stored.
func GetToken(configDir string) (*oauth2.Token, error) {
	creds, err := readCredentials(configDir)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  creds.AccessToken,
		TokenType:    creds.TokenType,
		RefreshToken: creds.RefreshToken,
		Expiry:       creds.Expiry,
	}, nil
}

// ProactiveRefresh refreshes the access token if it is valid but will expire
// within the given window. This is intended to be called after a command
//
//	completes, so the next invocation has a fresh token ready. Callers
//
// intentionally ignore errors — this is the best effort.
func ProactiveRefresh(configDir string, window time.Duration) error {
	if os.Getenv(EnvToken) != "" {
		return nil // env-supplied tokens are not managed by us
	}

	creds, err := readCredentials(configDir)
	if err != nil {
		return err
	}

	// Nothing to do if there's no expiry, no refresh token, or the token
	// isn't expiring within the window.
	if creds.Expiry.IsZero() || creds.RefreshToken == "" || creds.TokenURL == "" {
		return nil
	}
	remaining := time.Until(creds.Expiry)
	if remaining > window || remaining <= 0 {
		return nil // still fresh enough, or already expired (pre-command refresh handles that)
	}

	_, err = refreshCredentials(configDir, creds)
	return err
}

// DeleteToken removes the credentials file.
func DeleteToken(configDir string) error {
	path := filepath.Join(configDir, credentialsFile)
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func writeCredentials(configDir string, creds *credentials) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	path := filepath.Join(configDir, credentialsFile)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

func readCredentials(configDir string) (*credentials, error) {
	path := filepath.Join(configDir, credentialsFile)
	data, err := os.ReadFile(path) //nolint:gosec // path is constructed from configDir + constant filename
	if err != nil {
		return nil, err
	}

	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

// refreshCredentials uses the refresh token to obtain a new access token and persists it.
func refreshCredentials(configDir string, creds *credentials) (*credentials, error) {
	// Set Expiry to a past time so the oauth2 library considers the token
	// expired and actually performs the refresh. Without this, oauth2's
	// internal expiryDelta (10s) causes it to return the existing token
	// as-is when it still has >10s remaining — defeating our proactive
	// refresh window. Note: a zero Expiry means "never expires" in oauth2,
	// so we must use an explicitly past time.
	token := &oauth2.Token{
		AccessToken:  creds.AccessToken,
		TokenType:    creds.TokenType,
		RefreshToken: creds.RefreshToken,
		Expiry:       time.Now().Add(-time.Minute),
	}

	cfg := &oauth2.Config{
		ClientID: creds.ClientID,
		Endpoint: oauth2.Endpoint{TokenURL: creds.TokenURL},
	}

	refreshed, err := cfg.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, err
	}

	newCreds := &credentials{
		AccessToken:  refreshed.AccessToken,
		TokenType:    refreshed.TokenType,
		RefreshToken: refreshed.RefreshToken,
		Expiry:       refreshed.Expiry,
		ClientID:     creds.ClientID,
		TokenURL:     creds.TokenURL,
	}

	// Persist the refreshed token so we don't refresh again on the next call.
	if err := writeCredentials(configDir, newCreds); err != nil {
		return nil, err
	}

	return newCreds, nil
}
