package auth

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cli/browser"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"go.admiral.io/cli/internal/credentials"
)

//go:embed templates/login.gohtml
var resultPage []byte

//go:embed templates/error.gohtml
var errorPageRaw string

var errorPageTmpl = template.Must(template.New("error").Parse(errorPageRaw))

// LoginOptions configures the OIDC login flow.
type LoginOptions struct {
	Issuer    string
	ClientID  string
	Scopes    []string
	ConfigDir string
}

// callbackPorts is the set of ports pre-registered as redirect URIs on the
// auth server. The login flow tries each in order until one is available.
var callbackPorts = []int{1597, 2584, 4181, 6765, 10946, 17711, 28657, 46368}

// Login performs the OIDC browser-based login flow with PKCE.
func Login(ctx context.Context, opts LoginOptions) error {
	// Try each pre-registered port until we find one available. We bind to
	// 127.0.0.1 (loopback only, RFC 8252 8.3) but use localhost in the
	// redirect URI to match what's whitelisted on the auth server.
	var port int
	var ln net.Listener
	for _, p := range callbackPorts {
		var err error
		ln, err = net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", p))
		if err == nil {
			port = p
			break
		}
	}
	if ln == nil {
		return errors.New("unable to find an available port for the callback listener")
	}
	defer ln.Close() //nolint:errcheck // best-effort cleanup

	// OIDC discovery with a dedicated HTTP client (not http.DefaultClient).
	httpClient := &http.Client{Timeout: 30 * time.Second}
	oidcCtx := oidc.ClientContext(ctx, httpClient)
	provider, err := oidc.NewProvider(oidcCtx, opts.Issuer)
	if err != nil {
		return fmt.Errorf("failed to query OIDC provider %q: %w", opts.Issuer, err)
	}

	// Generate cryptographic state, nonce, and PKCE verifier.
	state, err := generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}
	nonce, err := generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	verifier := oauth2.GenerateVerifier()

	// Use localhost in the redirect URI for IdP compatibility even though
	// we bind to 127.0.0.1 (RFC 8252 §7.3).
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)
	oc := &oauth2.Config{
		ClientID:    opts.ClientID,
		Endpoint:    provider.Endpoint(),
		RedirectURL: redirectURL,
		Scopes:      opts.Scopes,
	}

	authCodeURL := oc.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("nonce", nonce),
	)

	// Channel to receive the authorization code or error from the callback.
	type authResult struct {
		code string
		err  error
	}
	resultCh := make(chan authResult, 1)
	var once sync.Once

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			renderError(w, http.StatusMethodNotAllowed, "Method not allowed.")
			once.Do(func() {
				resultCh <- authResult{err: errors.New("unexpected callback method: " + r.Method)}
			})
			return
		}

		once.Do(func() {
			q := r.URL.Query()

			// Check for IdP error first (e.g. user denied consent).
			if errCode := q.Get("error"); errCode != "" {
				desc := q.Get("error_description")
				renderError(w, http.StatusBadRequest, fmt.Sprintf("Authorization failed: %s", desc))
				resultCh <- authResult{err: fmt.Errorf("authorization denied: %s: %s", errCode, desc)}
				return
			}

			// Validate state to prevent CSRF.
			if q.Get("state") != state {
				renderError(w, http.StatusBadRequest, "State mismatch — please try logging in again.")
				resultCh <- authResult{err: errors.New("state mismatch")}
				return
			}

			code := q.Get("code")
			if code == "" {
				renderError(w, http.StatusBadRequest, "Missing authorization code.")
				resultCh <- authResult{err: errors.New("missing authorization code")}
				return
			}

			// Render the success page immediately; token exchange happens outside
			// the handler, so the browser gets its response without delay.
			_, _ = w.Write(resultPage)
			resultCh <- authResult{code: code}
		})
	})

	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	// Run the callback server in the background.
	go func() {
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			resultCh <- authResult{err: fmt.Errorf("callback server failed: %w", err)}
		}
	}()

	slog.Info("opening browser for authentication", "url", authCodeURL)
	if err := browser.OpenURL(authCodeURL); err != nil {
		gracefulShutdown(server)
		return fmt.Errorf("unable to open browser: %w", err)
	}

	// Wait for the callback or context cancellation (e.g. Ctrl+C).
	var res authResult
	select {
	case <-ctx.Done():
		gracefulShutdown(server)
		return fmt.Errorf("login canceled: %w", ctx.Err())
	case res = <-resultCh:
	}

	// Shut down gracefully so the success page is fully delivered.
	gracefulShutdown(server)

	if res.err != nil {
		return res.err
	}

	// Exchange the authorization code for tokens.
	token, err := oc.Exchange(oidcCtx, res.code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	// Verify ID token: signature, issuer, audience, and expiry (OIDC Core §3.1.3.7).
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return errors.New("missing id_token in token response")
	}

	idToken, err := provider.Verifier(&oidc.Config{ClientID: opts.ClientID}).Verify(ctx, rawIDToken)
	if err != nil {
		return fmt.Errorf("id token verification failed: %w", err)
	}

	if idToken.Nonce != nonce {
		return errors.New("nonce mismatch")
	}

	slog.Debug("login flow completed")
	return credentials.SaveToken(opts.ConfigDir, token, opts.ClientID, oc.Endpoint.TokenURL)
}

// renderError writes the styled error page with the given HTTP status and message.
func renderError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = errorPageTmpl.Execute(w, struct{ Message string }{msg})
}

// gracefulShutdown attempts a clean server shutdown, falling back to a hard
// close if it doesn't complete within 500ms.
func gracefulShutdown(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		_ = server.Close()
	}
}

// generateRandomString returns a URL-safe base64-encoded random string.
func generateRandomString(n int) (string, error) {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
