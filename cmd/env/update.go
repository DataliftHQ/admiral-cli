package env

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
)

func newUpdateCmd(opts *factory.Options) *cobra.Command {
	var (
		promotionOrder int32
		ttl            time.Duration
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an environment",
		Long:  `Update an existing environment.`,
		Example: `  # Update promotion order
  admiral use billing-api
  admiral env update production --promotion-order 100

  # Update TTL for ephemeral environment
  admiral env update preview-123 --ttl 48h`,
		Args: cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			app, err := resolveAppForEnv(opts.ConfigDir)
			if err != nil {
				return err
			}

			flags := map[string]any{}
			if promotionOrder != 0 {
				flags["promotion-order"] = promotionOrder
			}
			if ttl > 0 {
				flags["ttl"] = ttl.String()
			}

			stub := cmdutil.StubResult{
				Command:     "env update",
				App:         app,
				Environment: name,
				Flags:       flags,
				Status:      cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().Int32Var(&promotionOrder, "promotion-order", 0, "promotion order (lower = earlier)")
	cmd.Flags().DurationVar(&ttl, "ttl", 0, "time-to-live for ephemeral environments (e.g. 24h)")

	return cmd
}
