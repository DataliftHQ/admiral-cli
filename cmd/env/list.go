package env

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
)

func newListCmd(opts *factory.Options) *cobra.Command {
	var (
		pageSize  int32
		pageToken string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List environments",
		Long:  `List all environments for the active application.`,
		Example: `  # List environments
  admiral use billing-api
  admiral env list

  # Paginated listing
  admiral env list --page-size 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := resolveAppForEnv(opts.ConfigDir)
			if err != nil {
				return err
			}

			flags := map[string]any{}
			if pageSize > 0 {
				flags["page-size"] = pageSize
			}
			if pageToken != "" {
				flags["page-token"] = pageToken
			}

			stub := cmdutil.StubResult{
				Command: "env list",
				App:     app,
				Flags:   flags,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 0, "maximum number of results per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "pagination token from a previous response")

	return cmd
}
