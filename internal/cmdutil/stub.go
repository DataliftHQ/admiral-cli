package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"gopkg.in/yaml.v3"

	"go.admiral.io/cli/internal/output"
)

// StubResult holds the parameters that would be sent to the API.
type StubResult struct {
	Command     string            `json:"command" yaml:"command"`
	App         string            `json:"app,omitempty" yaml:"app,omitempty"`
	Environment string            `json:"environment,omitempty" yaml:"environment,omitempty"`
	Flags       map[string]any    `json:"flags,omitempty" yaml:"flags,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Status      string            `json:"status" yaml:"status"`
}

// StubStatus is the default status message for stub commands.
const StubStatus = "not yet implemented (proto/gRPC pending)"

// PrintStub writes a StubResult in the requested output format.
func PrintStub(w io.Writer, format output.Format, stub StubResult) error {
	switch format {
	case output.FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stub)
	case output.FormatYAML:
		return yaml.NewEncoder(w).Encode(stub)
	case output.FormatTable, output.FormatWide:
		return PrintStubTable(w, stub)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// PrintStubTable writes a StubResult as a human-readable table.
func PrintStubTable(out io.Writer, stub StubResult) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)

	output.Writef(w, "Command:\t%s\n", stub.Command)

	if stub.App != "" {
		output.Writef(w, "App:\t%s\n", stub.App)
	}
	if stub.Environment != "" {
		output.Writef(w, "Environment:\t%s\n", stub.Environment)
	}

	if len(stub.Flags) > 0 {
		output.Writeln(w, "Flags:")
		for k, v := range stub.Flags {
			output.Writef(w, "  %s\t= %v\n", k, v)
		}
	}

	if len(stub.Labels) > 0 {
		output.Writeln(w, "Labels:")
		for k, v := range stub.Labels {
			output.Writef(w, "  %s\t= %s\n", k, v)
		}
	}

	output.Writef(w, "\nStatus:\t%s\n", stub.Status)

	return w.Flush()
}
