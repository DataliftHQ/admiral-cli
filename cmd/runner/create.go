package runner

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	runnerv1 "go.admiral.io/sdk/proto/runner/v1"
)

func newCreateCmd(opts *factory.Options) *cobra.Command {
	var (
		name   string
		kind   string
		labels []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a runner",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedKind, err := parseRunnerKind(kind)
			if err != nil {
				return err
			}

			parsedLabels, err := parseLabels(labels)
			if err != nil {
				return err
			}

			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Runner().CreateRunner(cmd.Context(), &runnerv1.CreateRunnerRequest{
				DisplayName: name,
				Kind:        parsedKind,
				Labels:      parsedLabels,
			})
			if err != nil {
				return err
			}

			p := output.NewPrinter(opts.OutputFormat)
			if err := p.PrintResource(resp, func(w *tabwriter.Writer) {
				printRunnerTable(w, []*runnerv1.Runner{resp.Runner}, opts.OutputFormat)
			}); err != nil {
				return err
			}

			if resp.PlainTextToken != "" {
				output.PrintToken(cmd.ErrOrStderr(), resp.PlainTextToken)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "display name for the runner (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&kind, "kind", "", "runner kind: terraform or workflow (required)")
	_ = cmd.MarkFlagRequired("kind")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "labels in key=value format (can be repeated)")

	return cmd
}

func parseRunnerKind(s string) (runnerv1.RunnerKind, error) {
	switch strings.ToLower(s) {
	case "terraform":
		return runnerv1.RunnerKind_RUNNER_KIND_TERRAFORM, nil
	case "workflow":
		return runnerv1.RunnerKind_RUNNER_KIND_WORKFLOW, nil
	default:
		return 0, fmt.Errorf("invalid runner kind %q: must be terraform or workflow", s)
	}
}

func parseLabels(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	labels := make(map[string]string, len(raw))
	for _, l := range raw {
		k, v, ok := strings.Cut(l, "=")
		if !ok {
			return nil, fmt.Errorf("invalid label %q: must be in key=value format", l)
		}
		labels[k] = v
	}
	return labels, nil
}
