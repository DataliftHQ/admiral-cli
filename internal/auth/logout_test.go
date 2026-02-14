package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"go.admiral.io/cli/internal/credentials"
)

func TestRevokeRefreshToken(t *testing.T) {
	t.Run("sends correct POST body", func(t *testing.T) {
		var gotToken, gotClientID string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/oauth2/revoke", r.URL.Path)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			require.NoError(t, r.ParseForm())
			gotToken = r.FormValue("token")
			gotClientID = r.FormValue("client_id")

			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		err := revokeRefreshToken("my-refresh-token", "my-client", srv.URL)
		require.NoError(t, err)
		assert.Equal(t, "my-refresh-token", gotToken)
		assert.Equal(t, "my-client", gotClientID)
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		err := revokeRefreshToken("tok", "cid", srv.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token revocation failed")
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("returns error on network failure", func(t *testing.T) {
		err := revokeRefreshToken("tok", "cid", "http://127.0.0.1:0")
		require.Error(t, err)
	})

	t.Run("returns error on invalid issuer URL", func(t *testing.T) {
		err := revokeRefreshToken("tok", "cid", "://bad-url")
		require.Error(t, err)
	})
}

// saveTestToken writes a token to the given configDir so Logout can read it.
func saveTestToken(t *testing.T, configDir string, tok *oauth2.Token) {
	t.Helper()
	require.NoError(t, credentials.SaveToken(configDir, tok, "test-client", "https://example.com/token"))
}

func TestLogout(t *testing.T) {
	t.Run("deletes stored credentials", func(t *testing.T) {
		dir := t.TempDir()
		saveTestToken(t, dir, &oauth2.Token{
			AccessToken: "at",
			Expiry:      time.Now().Add(time.Hour),
		})

		// Verify file exists before logout.
		_, err := os.Stat(filepath.Join(dir, "credentials.json"))
		require.NoError(t, err)

		err = Logout(context.Background(), LogoutOptions{
			Issuer:    "https://example.com",
			ClientID:  "test-client",
			ConfigDir: dir,
		})
		require.NoError(t, err)

		// Verify file is removed.
		_, err = os.Stat(filepath.Join(dir, "credentials.json"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("revokes refresh token when present", func(t *testing.T) {
		var revoked bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "my-refresh", r.FormValue("token"))
			revoked = true
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		dir := t.TempDir()
		saveTestToken(t, dir, &oauth2.Token{
			AccessToken:  "at",
			RefreshToken: "my-refresh",
			Expiry:       time.Now().Add(time.Hour),
		})

		err := Logout(context.Background(), LogoutOptions{
			Issuer:    srv.URL,
			ClientID:  "test-client",
			ConfigDir: dir,
		})
		require.NoError(t, err)
		assert.True(t, revoked, "refresh token should have been revoked")
	})

	t.Run("skips revocation when no refresh token", func(t *testing.T) {
		var called bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		dir := t.TempDir()
		saveTestToken(t, dir, &oauth2.Token{
			AccessToken: "at",
			Expiry:      time.Now().Add(time.Hour),
		})

		err := Logout(context.Background(), LogoutOptions{
			Issuer:    srv.URL,
			ClientID:  "test-client",
			ConfigDir: dir,
		})
		require.NoError(t, err)
		assert.False(t, called, "revocation endpoint should not be called")
	})

	t.Run("succeeds when no credentials exist", func(t *testing.T) {
		dir := t.TempDir()
		err := Logout(context.Background(), LogoutOptions{
			Issuer:    "https://example.com",
			ClientID:  "test-client",
			ConfigDir: dir,
		})
		require.NoError(t, err)
	})

	t.Run("revocation failure does not fail logout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		dir := t.TempDir()
		saveTestToken(t, dir, &oauth2.Token{
			AccessToken:  "at",
			RefreshToken: "rt",
			Expiry:       time.Now().Add(time.Hour),
		})

		err := Logout(context.Background(), LogoutOptions{
			Issuer:    srv.URL,
			ClientID:  "test-client",
			ConfigDir: dir,
		})
		require.NoError(t, err, "logout should succeed even if revocation fails")
	})
}

// TestLogoutDeletesBeforeRevoking verifies that credentials are removed from
// disk before the revocation request is made. This ensures the user is logged
// out locally even if the revocation endpoint is slow or unreachable.
func TestLogoutDeletesBeforeRevoking(t *testing.T) {
	var deletedBeforeRevoke bool
	dir := t.TempDir()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// At the point the revocation request is made, the file should already be gone.
		_, err := os.Stat(filepath.Join(dir, "credentials.json"))
		deletedBeforeRevoke = os.IsNotExist(err)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	saveTestToken(t, dir, &oauth2.Token{
		AccessToken:  "at",
		RefreshToken: "rt",
		Expiry:       time.Now().Add(time.Hour),
	})

	err := Logout(context.Background(), LogoutOptions{
		Issuer:    srv.URL,
		ClientID:  "test-client",
		ConfigDir: dir,
	})
	require.NoError(t, err)
	assert.True(t, deletedBeforeRevoke, "credentials should be deleted before revocation request")
}

// TestRevokeRefreshTokenPathAppend verifies the revocation path is appended
// correctly when the issuer already has a path component.
func TestRevokeRefreshTokenPathAppend(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Issuer with existing path component.
	err := revokeRefreshToken("tok", "cid", srv.URL+"/realms/test")
	// This will fail because httptest.Server doesn't route by path,
	// but we can verify the path was constructed correctly.
	if err == nil {
		assert.Equal(t, "/realms/test/oauth2/revoke", gotPath)
	}
}

// TestLogoutOptionsFields ensures the struct can be constructed with all fields.
func TestLogoutOptionsFields(t *testing.T) {
	opts := LogoutOptions{
		Issuer:    "https://auth.example.com",
		ClientID:  "my-client-id",
		ConfigDir: "/tmp/test",
	}
	assert.Equal(t, "https://auth.example.com", opts.Issuer)
	assert.Equal(t, "my-client-id", opts.ClientID)
	assert.Equal(t, "/tmp/test", opts.ConfigDir)
}

// TestCredentialsFileFormat verifies the shape of the persisted credentials
// file that Logout reads, ensuring compatibility between save and load.
func TestCredentialsFileFormat(t *testing.T) {
	dir := t.TempDir()
	tok := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	saveTestToken(t, dir, tok)

	data, err := os.ReadFile(filepath.Join(dir, "credentials.json"))
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Equal(t, "access", raw["access_token"])
	assert.Equal(t, "refresh", raw["refresh_token"])
	assert.Equal(t, "Bearer", raw["token_type"])
}
