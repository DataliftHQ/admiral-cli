package app

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
		labels    []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List applications",
		Long:  `List all applications visible to the current user.`,
		Example: `  # List all applications
  admiral app list

  # List with label filter
  admiral app list --label team=platform

  # Paginated listing
  admiral app list --page-size 10`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := map[string]any{}
			if pageSize > 0 {
				flags["page-size"] = pageSize
			}
			if pageToken != "" {
				flags["page-token"] = pageToken
			}

			labelMap, err := cmdutil.ParseLabels(labels)
			if err != nil {
				return err
			}

			stub := cmdutil.StubResult{
				Command: "app list",
				Labels:  labelMap,
				Flags:   flags,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().Int32Var(&pageSize, "page-size", 0, "maximum number of results per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "pagination token from a previous response")
	cmdutil.AddLabelFlag(cmd, &labels, "filter by label (key=value, repeatable)")

	return cmd
}
