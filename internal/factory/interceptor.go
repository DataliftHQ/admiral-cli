package factory

import (
	"context"
	"log/slog"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	admiralcreds "go.admiral.io/cli/internal/credentials"
	"go.admiral.io/sdk/client"
)

// authRetryInterceptor returns a gRPC unary interceptor that retries once on
// Unauthenticated (401) after force-refreshing the stored token. The retry is
// limited to a single attempt per client lifetime via an atomic flag.
func authRetryInterceptor(configDir string, authScheme client.AuthScheme, insecure bool) grpc.UnaryClientInterceptor {
	var retried atomic.Bool

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}

		if status.Code(err) != codes.Unauthenticated {
			return err
		}

		// Only retry once per client lifetime.
		if !retried.CompareAndSwap(false, true) {
			return err
		}

		slog.Debug("received 401, attempting token refresh", "method", method)

		result, refreshErr := admiralcreds.ForceRefresh(configDir)
		if refreshErr != nil {
			slog.Debug("token refresh failed", "error", refreshErr)
			return err // return original 401 error
		}

		slog.Debug("token refreshed, retrying request", "method", method)

		retryOpts := append([]grpc.CallOption{}, opts...)
		retryOpts = append(retryOpts, grpc.PerRPCCredsCallOption{
			Creds: tokenCallCredentials{
				token:    result.Token,
				scheme:   authScheme,
				insecure: insecure,
			},
		})

		return invoker(ctx, method, req, reply, cc, retryOpts...)
	}
}

// tokenCallCredentials implements credentials.PerRPCCredentials for passing a
// fresh token as a gRPC call option on retry.
type tokenCallCredentials struct {
	token    string
	scheme   client.AuthScheme
	insecure bool
}

func (t tokenCallCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": t.scheme.String() + " " + t.token,
	}, nil
}

func (t tokenCallCredentials) RequireTransportSecurity() bool {
	return !t.insecure
}

// Compile-time check.
var _ credentials.PerRPCCredentials = tokenCallCredentials{}
