package env

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <slug>",
		Short: "Get environment details",
		Long:  `Get detailed information about an environment.`,
		Example: `  # Get environment details
  admiral use billing-api
  admiral env get staging`,
		Args: cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			app, err := resolveAppForEnv(opts.ConfigDir)
			if err != nil {
				return err
			}

			stub := cmdutil.StubResult{
				Command:     "env get",
				App:         app,
				Environment: slug,
				Status:      cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	return cmd
}
