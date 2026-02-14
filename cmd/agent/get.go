package agent

import (
	"github.com/spf13/cobra"

	"go.admiral.io/cli/internal/factory"
	"go.admiral.io/cli/internal/output"
	agentv1 "go.admiral.io/sdk/proto/agent/v1"
)

func newGetCmd(opts *factory.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <agent-id>",
		Short: "Get an agent by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := factory.CreateClient(cmd.Context(), opts)
			if err != nil {
				return err
			}
			defer c.Close() //nolint:errcheck // best-effort cleanup

			resp, err := c.Agent().GetAgent(cmd.Context(), &agentv1.GetAgentRequest{
				AgentId: args[0],
			})
			if err != nil {
				return err
			}

			a := resp.Agent
			p := output.NewPrinter(opts.OutputFormat)

			sections := []output.Section{
				{
					Details: []output.Detail{
						{Key: "Name", Value: a.DisplayName},
						{Key: "ID", Value: a.Id},
						{Key: "Status", Value: output.FormatEnum(a.Status.String(), "AGENT_STATUS_")},
						{Key: "Version", Value: a.Version},
						{Key: "Cluster", Value: a.ClusterId},
						{Key: "Runner", Value: a.RunnerId},
						{Key: "Last Heartbeat", Value: output.FormatTimestamp(a.LastHeartbeatAt)},
						{Key: "Created", Value: output.FormatTimestamp(a.CreatedAt)},
						{Key: "Updated", Value: output.FormatTimestamp(a.UpdatedAt)},
						{Key: "Age", Value: output.FormatAge(a.CreatedAt)},
					},
				},
			}

			return p.PrintDetail(resp, sections)
		},
	}
}
