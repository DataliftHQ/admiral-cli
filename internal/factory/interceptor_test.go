package factory

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.admiral.io/sdk/client"
)

// fakeInvoker returns a grpc.UnaryInvoker that returns the given errors in
// sequence. After exhausting the list it returns nil.
func fakeInvoker(errs ...error) grpc.UnaryInvoker {
	idx := 0
	return func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		if idx >= len(errs) {
			return nil
		}
		err := errs[idx]
		idx++
		return err
	}
}

// writeTestCredentials writes a credentials.json that has a refresh token and
// token URL so ForceRefresh can attempt a refresh (it will fail at the HTTP
// level, but that's fine — we test the interceptor's control flow, not the
// actual OAuth exchange).
func writeTestCredentials(t *testing.T, dir string) {
	t.Helper()

	creds := map[string]any{
		"access_token":  "old-token",
		"refresh_token": "rt",
		"client_id":     "cid",
		"token_url":     "https://auth.example.com/token",
		"expiry":        time.Now().Add(1 * time.Hour).Format(time.RFC3339),
	}

	data, err := json.Marshal(creds)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600))
}

func TestAuthRetryInterceptor_NoErrorPassThrough(t *testing.T) {
	interceptor := authRetryInterceptor(t.TempDir(), client.AuthSchemeBearer, true)

	err := interceptor(
		context.Background(), "/test.Service/Method",
		nil, nil, nil,
		fakeInvoker(nil),
	)
	require.NoError(t, err)
}

func TestAuthRetryInterceptor_NonAuthErrorPassThrough(t *testing.T) {
	interceptor := authRetryInterceptor(t.TempDir(), client.AuthSchemeBearer, true)

	notFound := status.Error(codes.NotFound, "not found")
	err := interceptor(
		context.Background(), "/test.Service/Method",
		nil, nil, nil,
		fakeInvoker(notFound),
	)
	require.Equal(t, notFound, err)
}

func TestAuthRetryInterceptor_RetriesOnceOn401(t *testing.T) {
	// The interceptor should attempt a retry on Unauthenticated. The retry
	// will also fail because we don't have a real OAuth server, but we verify
	// the invoker is called twice and the original 401 is returned (since
	// ForceRefresh fails without a real token endpoint).
	dir := t.TempDir()
	writeTestCredentials(t, dir)
	os.Unsetenv("ADMIRAL_TOKEN")

	callCount := 0
	invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		callCount++
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	interceptor := authRetryInterceptor(dir, client.AuthSchemeBearer, true)
	err := interceptor(
		context.Background(), "/test.Service/Method",
		nil, nil, nil,
		invoker,
	)

	// ForceRefresh will fail (no real OAuth server), so the original error is returned.
	require.Error(t, err)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
	// Invoker called exactly once — the retry didn't happen because refresh failed.
	require.Equal(t, 1, callCount)
}

func TestAuthRetryInterceptor_DoesNotRetryTwice(t *testing.T) {
	// Even if the first 401 triggers a retry attempt, a second 401 from a
	// separate call must NOT trigger another retry.
	dir := t.TempDir()
	os.Unsetenv("ADMIRAL_TOKEN")

	interceptor := authRetryInterceptor(dir, client.AuthSchemeBearer, true)

	unauthErr := status.Error(codes.Unauthenticated, "unauthenticated")

	// First call — will attempt refresh (and fail), consuming the one-time retry.
	_ = interceptor(
		context.Background(), "/test.Service/Method1",
		nil, nil, nil,
		fakeInvoker(unauthErr),
	)

	// Second call — should NOT attempt refresh again.
	callCount := 0
	invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		callCount++
		return unauthErr
	}

	err := interceptor(
		context.Background(), "/test.Service/Method2",
		nil, nil, nil,
		invoker,
	)

	require.Error(t, err)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
	// Invoker called exactly once — no retry attempted.
	require.Equal(t, 1, callCount)
}

func TestTokenCallCredentials_GetRequestMetadata(t *testing.T) {
	creds := tokenCallCredentials{
		token:    "my-token",
		scheme:   client.AuthSchemeBearer,
		insecure: true,
	}

	md, err := creds.GetRequestMetadata(context.Background())
	require.NoError(t, err)
	require.Equal(t, "Bearer my-token", md["authorization"])
}

func TestTokenCallCredentials_RequireTransportSecurity(t *testing.T) {
	secure := tokenCallCredentials{insecure: false}
	require.True(t, secure.RequireTransportSecurity())

	insecure := tokenCallCredentials{insecure: true}
	require.False(t, insecure.RequireTransportSecurity())
}
