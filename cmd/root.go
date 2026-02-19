package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/cobra"

	appcmd "go.admiral.io/cli/cmd/app"
	authcmd "go.admiral.io/cli/cmd/auth"
	clustercmd "go.admiral.io/cli/cmd/cluster"
	envcmd "go.admiral.io/cli/cmd/env"
	variablecmd "go.admiral.io/cli/cmd/variable"
	"go.admiral.io/cli/internal/config"
	"go.admiral.io/cli/internal/credentials"
	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	"go.admiral.io/cli/internal/version"
)

func Execute(version version.Version, exit func(int), args []string) {
	newRootCmd(version, exit).Execute(args)
}

func (cmd *rootCmd) Execute(args []string) {
	cmd.cmd.SetArgs(args)

	if err := cmd.cmd.Execute(); err != nil {
		code := 1
		eerr := &exitError{}
		if errors.As(err, &eerr) {
			code = eerr.code
		}
		output.Writef(os.Stderr, "Error: %s\n", err)
		cmd.exit(code)
	}
}

type rootCmd struct {
	cmd  *cobra.Command
	exit func(int)

	verbose      bool
	configPath   string
	outputFormat string
	factoryOpts  *factory.Options
}

func newRootCmd(ver version.Version, exit func(int)) *rootCmd {
	var factoryOpts factory.Options

	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:           "admiral",
		Short:         "Admiral - Platform Orchestrator",
		Version:       ver.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if root.verbose {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
				slog.Debug("debug logs enabled")
			}

			f, err := output.ParseFormat(root.outputFormat)
			if err != nil {
				return err
			}
			factoryOpts.OutputFormat = f
			factoryOpts.ConfigDir = root.configPath
			factoryOpts.Verbose = root.verbose

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Best-effort proactive token refresh: if the access token is
			// still valid but expiring soon, refresh it now so the next
			// invocation starts with a fresh token and avoids the refresh lag.
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			done := make(chan struct{})
			go func() {
				_ = credentials.ProactiveRefresh(root.configPath, time.Minute)
				close(done)
			}()

			select {
			case <-done:
			case <-ctx.Done():
				slog.Debug("proactive token refresh timed out")
			}
		},
	}
	cmd.SetVersionTemplate("{{.Version}}")

	defaultConfigPath, err := config.ConfigDir()
	if err != nil {
		slog.Error("failed to get default config path", "error", err)
		os.Exit(1)
	}

	// Config flags
	cmd.PersistentFlags().StringVar(&root.configPath, "config-dir", defaultConfigPath, "path to config directory")

	// Server flags
	cmd.PersistentFlags().StringVarP(&factoryOpts.ServerAddr, "server", "s", "api.admiral.io:443", "host:port of the API server")
	cmd.PersistentFlags().BoolVar(&factoryOpts.PlainText, "plaintext", false, "disable TLS")
	cmd.PersistentFlags().BoolVarP(&factoryOpts.Insecure, "insecure", "i", false, "skip server certificate and domain verification")

	// Output flags
	cmd.PersistentFlags().StringVarP(&root.outputFormat, "output", "o", "table", "output format: table, json, yaml, wide")

	// Auth flags (hidden, for dev/testing override)
	cmd.PersistentFlags().StringVar(&factoryOpts.Issuer, "issuer", "https://auth.admiral.io", "OIDC identity provider URL")
	_ = cmd.PersistentFlags().MarkHidden("issuer")
	cmd.PersistentFlags().StringVar(&factoryOpts.ClientID, "client-id", "44972cb5-9739-4b8d-ac29-6dcccca3e9db", "OAuth2 client ID")
	_ = cmd.PersistentFlags().MarkHidden("client-id")
	cmd.PersistentFlags().StringSliceVar(&factoryOpts.Scopes, "scopes", []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess}, "OAuth2 scopes")
	_ = cmd.PersistentFlags().MarkHidden("scopes")

	// General flags
	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false, "enable verbose mode")
	cmd.PersistentFlags().BoolP("help", "h", false, "help for admiral")

	// Auth commands
	authCmd := authcmd.NewAuthCmd(&factoryOpts)
	cmd.AddCommand(authCmd.Cmd)

	// Resource commands
	cmd.AddCommand(
		appcmd.NewAppCmd(&factoryOpts).Cmd,
		clustercmd.NewClusterCmd(&factoryOpts).Cmd,
		envcmd.NewEnvCmd(&factoryOpts).Cmd,
		variablecmd.NewVariableCmd(&factoryOpts).Cmd,
	)

	// Utility commands
	cmd.AddCommand(
		newCompletionCmd(),
		newUseCmd(&factoryOpts),
		newVersionCmd(ver),
		newWhoamiCmd(&factoryOpts),
	)

	root.cmd = cmd
	root.factoryOpts = &factoryOpts

	return root
}
