package app

import (
	"os"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/cmdutil"
	"go.admiral.io/cli/internal/factory"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		name   string
		labels []string
	)

	cmd := &cobra.Command{
		Use:   "create <slug>",
		Short: "Create an application",
		Long:  `Create a new application with the given slug.`,
		Example: `  # Create an application
  admiral app create billing-api

  # Create with a display name and labels
  admiral app create billing-api --name "Billing API" --label team=platform`,
		Args: cmdutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			labelMap, err := cmdutil.ParseLabels(labels)
			if err != nil {
				return err
			}

			flags := map[string]any{}
			if name != "" {
				flags["name"] = name
			}

			stub := cmdutil.StubResult{
				Command: "app create",
				App:     slug,
				Labels:  labelMap,
				Flags:   flags,
				Status:  cmdutil.StubStatus,
			}

			return cmdutil.PrintStub(os.Stdout, opts.OutputFormat, stub)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the application")
	cmdutil.AddLabelFlag(cmd, &labels, "label to attach (key=value, repeatable)")

	return cmd
}
