package factory

import (
	"context"
	"fmt"
	"log/slog"

	"go.admiral.io/cli/internal/credentials"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/sdk/client"
)

// Options hold the configuration shared across all commands.
type Options struct {
	ServerAddr   string
	Insecure     bool
	PlainText    bool
	Verbose      bool
	ConfigDir    string
	OutputFormat output.Format

	// OIDC settings (used by auth commands)
	Issuer   string
	ClientID string
	Scopes   []string
}

// CreateClient creates a new AdmiralClient using the SDK.
func CreateClient(_ context.Context, opts *Options) (client.AdmiralClient, error) {
	result, err := credentials.ResolveToken(opts.ConfigDir)
	if err != nil {
		return nil, err
	}

	cfg := client.Config{
		HostPort:   opts.ServerAddr,
		AuthToken:  result.Token,
		AuthScheme: result.AuthScheme,
		ConnectionOptions: client.ConnectionOptions{
			Insecure: opts.Insecure || opts.PlainText,
		},
	}

	if opts.Verbose {
		cfg.Logger = client.NewSlogLogger(slog.Default())
	}

	c, err := client.New(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", opts.ServerAddr, err)
	}

	return c, nil
}
