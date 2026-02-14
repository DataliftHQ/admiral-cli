package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomString(t *testing.T) {
	t.Run("returns base64url encoded string of expected length", func(t *testing.T) {
		s, err := generateRandomString(32)
		require.NoError(t, err)

		// 32 raw bytes → 43 base64url chars (no padding).
		assert.Len(t, s, 43)

		// Must decode back to exactly 32 bytes.
		decoded, err := base64.RawURLEncoding.DecodeString(s)
		require.NoError(t, err)
		assert.Len(t, decoded, 32)
	})

	t.Run("produces unique values", func(t *testing.T) {
		a, err := generateRandomString(32)
		require.NoError(t, err)
		b, err := generateRandomString(32)
		require.NoError(t, err)

		assert.NotEqual(t, a, b)
	})
}

func TestRenderError(t *testing.T) {
	t.Run("sets status code and content type", func(t *testing.T) {
		w := httptest.NewRecorder()
		renderError(w, http.StatusBadRequest, "something went wrong")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	})

	t.Run("renders error message in HTML body", func(t *testing.T) {
		w := httptest.NewRecorder()
		renderError(w, http.StatusBadRequest, "token expired")

		body := w.Body.String()
		assert.Contains(t, body, "token expired")
		assert.Contains(t, body, "Authentication failed")
		assert.Contains(t, body, "<!DOCTYPE html>")
	})

	t.Run("escapes HTML in message", func(t *testing.T) {
		w := httptest.NewRecorder()
		renderError(w, http.StatusBadRequest, `<script>alert("xss")</script>`)

		body := w.Body.String()
		assert.NotContains(t, body, "<script>")
		assert.Contains(t, body, "&lt;script&gt;")
	})

	t.Run("supports different status codes", func(t *testing.T) {
		w := httptest.NewRecorder()
		renderError(w, http.StatusMethodNotAllowed, "Method not allowed.")

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestTemplatesEmbedded(t *testing.T) {
	t.Run("success page is embedded and non-empty", func(t *testing.T) {
		require.NotEmpty(t, resultPage, "login.gohtml should be embedded")
		assert.True(t, strings.Contains(string(resultPage), "Authentication successful"))
	})

	t.Run("error template is parsed and non-nil", func(t *testing.T) {
		require.NotNil(t, errorPageTmpl, "error.gohtml template should be parsed")
	})
}
