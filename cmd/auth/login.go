package auth

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	internalauth "go.admiral.io/cli/internal/auth"
	"go.admiral.io/cli/internal/factory"
)

// NewLoginCmd creates the login command.
func NewLoginCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in to Admiral",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := internalauth.Login(context.Background(), internalauth.LoginOptions{
				Issuer:    opts.Issuer,
				ClientID:  opts.ClientID,
				Scopes:    opts.Scopes,
				ConfigDir: opts.ConfigDir,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Successfully logged in.")
			return nil
		},
	}
}
