package env

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		cluster        string
		promotionOrder int32
		lifecycle      string
		parent         string
		sourceRef      string
		ttl            time.Duration
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an environment",
		Long:  `Create a new environment within the active application.`,
		Example: `  # Create a permanent environment
  admiral use billing-api
  admiral env create staging --cluster us-east-1

  # Create an ephemeral environment with TTL
  admiral env create preview-123 --cluster us-east-1 --lifecycle ephemeral --ttl 24h

  # Create with promotion order
  admiral env create production --cluster us-east-1 --promotion-order 100`,
		Args: cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			app, err := resolveAppForEnv(opts.ConfigDir)
			if err != nil {
				return err
			}

			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}

			if ttl > 0 && lifecycle != "ephemeral" {
				return fmt.Errorf("--ttl can only be used with --lifecycle ephemeral")
			}

			flags := map[string]any{
				"cluster": cluster,
			}
			if promotionOrder != 0 {
				flags["promotion-order"] = promotionOrder
			}
			if lifecycle != "" {
				flags["lifecycle"] = lifecycle
			}
			if parent != "" {
				flags["parent"] = parent
			}
			if sourceRef != "" {
				flags["source-ref"] = sourceRef
			}
			if ttl > 0 {
				flags["ttl"] = ttl.String()
			}

			stub := cmdutil.StubResult{
				Command:     "env create",
				App:         app,
				Environment: name,
				Flags:       flags,
				Status:      cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "target cluster (required)")
	cmd.Flags().Int32Var(&promotionOrder, "promotion-order", 0, "promotion order (lower = earlier)")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "lifecycle type: permanent or ephemeral")
	cmd.Flags().StringVar(&parent, "parent", "", "parent environment name")
	cmd.Flags().StringVar(&sourceRef, "source-ref", "", "source reference (branch, tag, or commit)")
	cmd.Flags().DurationVar(&ttl, "ttl", 0, "time-to-live for ephemeral environments (e.g. 24h)")

	return cmd
}
