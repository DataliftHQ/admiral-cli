package credentials

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"go.admiral.io/sdk/client"
)

func TestSaveAndGetToken(t *testing.T) {
	dir := t.TempDir()

	token := &oauth2.Token{
		AccessToken:  "access-123",
		TokenType:    "Bearer",
		RefreshToken: "refresh-456",
		Expiry:       time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := SaveToken(dir, token, "client-id", "https://auth.example.com/token"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	// Verify file permissions.
	info, err := os.Stat(filepath.Join(dir, credentialsFile))
	if err != nil {
		t.Fatalf("stat credentials file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("expected file mode 0600, got %o", perm)
	}

	got, err := GetToken(dir)
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if got.AccessToken != token.AccessToken {
		t.Fatalf("AccessToken: want %q, got %q", token.AccessToken, got.AccessToken)
	}
	if got.TokenType != token.TokenType {
		t.Fatalf("TokenType: want %q, got %q", token.TokenType, got.TokenType)
	}
	if got.RefreshToken != token.RefreshToken {
		t.Fatalf("RefreshToken: want %q, got %q", token.RefreshToken, got.RefreshToken)
	}
	if !got.Expiry.Equal(token.Expiry) {
		t.Fatalf("Expiry: want %v, got %v", token.Expiry, got.Expiry)
	}
}

func TestSaveToken_CreatesConfigDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "config")

	token := &oauth2.Token{AccessToken: "tok"}
	if err := SaveToken(dir, token, "cid", "https://auth.example.com/token"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, credentialsFile)); err != nil {
		t.Fatalf("credentials file not created: %v", err)
	}
}

func TestGetToken_NoFile(t *testing.T) {
	dir := t.TempDir()

	_, err := GetToken(dir)
	if err == nil {
		t.Fatal("expected error when no credentials file exists")
	}
}

func TestDeleteToken(t *testing.T) {
	dir := t.TempDir()

	token := &oauth2.Token{AccessToken: "to-delete"}
	if err := SaveToken(dir, token, "cid", "url"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	if err := DeleteToken(dir); err != nil {
		t.Fatalf("DeleteToken: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, credentialsFile)); !os.IsNotExist(err) {
		t.Fatal("expected credentials file to be deleted")
	}
}

func TestDeleteToken_NoFile(t *testing.T) {
	dir := t.TempDir()

	// Should not error when file doesn't exist.
	if err := DeleteToken(dir); err != nil {
		t.Fatalf("DeleteToken on missing file: %v", err)
	}
}

func TestResolveToken_EnvVar(t *testing.T) {
	t.Setenv(EnvToken, "env-token-789")

	got, err := ResolveToken(t.TempDir())
	require.NoError(t, err)
	require.Equal(t, "env-token-789", got.Token)
	require.Equal(t, client.AuthSchemeToken, got.AuthScheme)
}

func TestResolveToken_EnvVarTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	token := &oauth2.Token{
		AccessToken: "file-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	require.NoError(t, SaveToken(dir, token, "cid", "url"))

	t.Setenv(EnvToken, "env-token")

	got, err := ResolveToken(dir)
	require.NoError(t, err)
	require.Equal(t, "env-token", got.Token)
	require.Equal(t, client.AuthSchemeToken, got.AuthScheme)
}

func TestResolveToken_ValidToken(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	token := &oauth2.Token{
		AccessToken: "valid-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	require.NoError(t, SaveToken(dir, token, "cid", "url"))

	got, err := ResolveToken(dir)
	require.NoError(t, err)
	require.Equal(t, "valid-token", got.Token)
	require.Equal(t, client.AuthSchemeBearer, got.AuthScheme)
}

func TestResolveToken_NoExpiry(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	// Zero expiry should be treated as valid (e.g. long-lived tokens).
	token := &oauth2.Token{AccessToken: "no-expiry-token"}
	require.NoError(t, SaveToken(dir, token, "cid", "url"))

	got, err := ResolveToken(dir)
	require.NoError(t, err)
	require.Equal(t, "no-expiry-token", got.Token)
	require.Equal(t, client.AuthSchemeBearer, got.AuthScheme)
}

func TestResolveToken_NotLoggedIn(t *testing.T) {
	os.Unsetenv(EnvToken)

	_, err := ResolveToken(t.TempDir())
	if err == nil {
		t.Fatal("expected error when not logged in")
	}
	if want := "not logged in"; !containsStr(err.Error(), want) {
		t.Fatalf("error should contain %q, got %q", want, err.Error())
	}
}

func TestResolveToken_EmptyAccessToken(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	// Write credentials with empty access token.
	creds := &credentials{AccessToken: ""}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	_, err := ResolveToken(dir)
	if err == nil {
		t.Fatal("expected error for empty access token")
	}
	if want := "not logged in"; !containsStr(err.Error(), want) {
		t.Fatalf("error should contain %q, got %q", want, err.Error())
	}
}

func TestResolveToken_ExpiredNoRefreshToken(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	creds := &credentials{
		AccessToken: "expired-token",
		Expiry:      time.Now().Add(-1 * time.Minute),
	}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	_, err := ResolveToken(dir)
	if err == nil {
		t.Fatal("expected error for expired token without refresh token")
	}
	if want := "session expired"; !containsStr(err.Error(), want) {
		t.Fatalf("error should contain %q, got %q", want, err.Error())
	}
}

func TestProactiveRefresh_EnvVarSkips(t *testing.T) {
	t.Setenv(EnvToken, "env-token")

	// Should return nil immediately when env var is set.
	if err := ProactiveRefresh(t.TempDir(), 2*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProactiveRefresh_NoCredentialsFile(t *testing.T) {
	os.Unsetenv(EnvToken)

	err := ProactiveRefresh(t.TempDir(), 2*time.Minute)
	if err == nil {
		t.Fatal("expected error when no credentials file exists")
	}
}

func TestProactiveRefresh_NoExpiry(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	creds := &credentials{AccessToken: "tok", RefreshToken: "rt", TokenURL: "url"}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	// Zero expiry — nothing to refresh.
	if err := ProactiveRefresh(dir, 2*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProactiveRefresh_TokenStillFresh(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	creds := &credentials{
		AccessToken:  "fresh-token",
		RefreshToken: "rt",
		TokenURL:     "url",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	// Token expires in 1h, window is 2min — should skip.
	if err := ProactiveRefresh(dir, 2*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProactiveRefresh_TokenAlreadyExpired(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	creds := &credentials{
		AccessToken:  "expired",
		RefreshToken: "rt",
		TokenURL:     "url",
		Expiry:       time.Now().Add(-1 * time.Minute),
	}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	// Already expired — pre-command refresh handles this, not proactive.
	if err := ProactiveRefresh(dir, 2*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProactiveRefresh_MissingRefreshToken(t *testing.T) {
	dir := t.TempDir()
	os.Unsetenv(EnvToken)

	creds := &credentials{
		AccessToken: "tok",
		TokenURL:    "url",
		Expiry:      time.Now().Add(30 * time.Second),
	}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	// Within window but no refresh token — should skip.
	if err := ProactiveRefresh(dir, 2*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteReadCredentials_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	original := &credentials{
		AccessToken:  "at",
		TokenType:    "Bearer",
		RefreshToken: "rt",
		Expiry:       time.Date(2030, 6, 15, 12, 0, 0, 0, time.UTC),
		ClientID:     "cid",
		TokenURL:     "https://auth.example.com/token",
	}

	if err := writeCredentials(dir, original); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	got, err := readCredentials(dir)
	if err != nil {
		t.Fatalf("readCredentials: %v", err)
	}

	if got.AccessToken != original.AccessToken ||
		got.TokenType != original.TokenType ||
		got.RefreshToken != original.RefreshToken ||
		got.ClientID != original.ClientID ||
		got.TokenURL != original.TokenURL ||
		!got.Expiry.Equal(original.Expiry) {
		t.Fatalf("round-trip mismatch:\n  want: %+v\n  got:  %+v", original, got)
	}
}

func TestReadCredentials_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, credentialsFile)

	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := readCredentials(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWriteCredentials_JSONFormat(t *testing.T) {
	dir := t.TempDir()

	creds := &credentials{
		AccessToken: "tok",
		ClientID:    "cid",
	}
	if err := writeCredentials(dir, creds); err != nil {
		t.Fatalf("writeCredentials: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, credentialsFile))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Verify it's valid, indented JSON.
	if !json.Valid(data) {
		t.Fatal("credentials file is not valid JSON")
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if parsed["access_token"] != "tok" {
		t.Fatalf("expected access_token=tok, got %v", parsed["access_token"])
	}
}

func TestCheckFilePermissions_SecureFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, credentialsFile)
	require.NoError(t, os.WriteFile(path, []byte(`{}`), 0600))

	// Should not panic or error — just a no-op for secure files.
	checkFilePermissions(path)
}

func TestCheckFilePermissions_InsecureFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, credentialsFile)
	require.NoError(t, os.WriteFile(path, []byte(`{}`), 0644))

	// Should not panic — just logs a warning.
	checkFilePermissions(path)

	// Verify the file is indeed too open.
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.NotEqual(t, os.FileMode(0600), info.Mode().Perm())
}

func TestCheckFilePermissions_MissingFile(t *testing.T) {
	// Should not panic when file doesn't exist.
	checkFilePermissions(filepath.Join(t.TempDir(), "nonexistent"))
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
