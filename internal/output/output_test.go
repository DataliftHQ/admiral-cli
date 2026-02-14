package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ---------------------------------------------------------------------------
// format.go
// ---------------------------------------------------------------------------

func TestParseFormat(t *testing.T) {
	valid := []struct {
		input string
		want  Format
	}{
		{"table", FormatTable},
		{"json", FormatJSON},
		{"yaml", FormatYAML},
		{"wide", FormatWide},
	}
	for _, tc := range valid {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseFormat(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}

	t.Run("invalid", func(t *testing.T) {
		_, err := ParseFormat("xml")
		if err == nil {
			t.Fatal("expected error for invalid format")
		}
		if !strings.Contains(err.Error(), "invalid output format") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, err := ParseFormat("")
		if err == nil {
			t.Fatal("expected error for empty format")
		}
	})
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatTable, "table"},
		{FormatJSON, "json"},
		{FormatYAML, "yaml"},
		{FormatWide, "wide"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.format.String(); got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// printer.go
// ---------------------------------------------------------------------------

func TestNewPrinter(t *testing.T) {
	p := NewPrinter(FormatJSON)
	if p.Format != FormatJSON {
		t.Fatalf("want format %q, got %q", FormatJSON, p.Format)
	}
	if p.Out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestPrintResource_Table(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Out: &buf}

	err := p.PrintResource(structpb.NewStringValue("test"), func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "NAME\tAGE")
		fmt.Fprintln(w, "foo\t5d")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "foo") {
		t.Fatalf("unexpected table output: %q", out)
	}
}

func TestPrintResource_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Out: &buf}

	msg := structpb.NewStringValue("hello")
	err := p.PrintResource(msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("output is not valid JSON: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("expected 'hello' in JSON output: %q", buf.String())
	}
}

func TestPrintResource_YAML(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatYAML, Out: &buf}

	msg := structpb.NewStringValue("world")
	err := p.PrintResource(msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "world") {
		t.Fatalf("expected 'world' in YAML output: %q", buf.String())
	}
}

func TestPrintResource_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "xml", Out: &buf}

	err := p.PrintResource(structpb.NewStringValue("test"), nil)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrintDetail_Table(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Out: &buf}

	sections := []Section{
		{
			Details: []Detail{
				{Key: "Name", Value: "prod-cluster"},
				{Key: "ID", Value: "abc-123"},
			},
		},
		{
			Name: "Labels",
			Details: []Detail{
				{Key: "env", Value: "production"},
			},
		},
	}

	err := p.PrintDetail(structpb.NewStringValue("test"), sections)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Name:") || !strings.Contains(out, "prod-cluster") {
		t.Fatalf("expected Name detail in output: %q", out)
	}
	if !strings.Contains(out, "Labels:") {
		t.Fatalf("expected Labels section header in output: %q", out)
	}
	if !strings.Contains(out, "  env:") {
		t.Fatalf("expected indented detail under section: %q", out)
	}
}

func TestPrintDetail_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Out: &buf}

	msg := structpb.NewStringValue("detail-test")
	err := p.PrintDetail(msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("output is not valid JSON: %q", buf.String())
	}
}

func TestPrintDetail_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "xml", Out: &buf}

	err := p.PrintDetail(structpb.NewStringValue("test"), nil)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestPrintToken(t *testing.T) {
	var buf bytes.Buffer
	PrintToken(&buf, "secret-token-xyz")

	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Fatalf("expected WARNING in output: %q", out)
	}
	if !strings.Contains(out, "secret-token-xyz") {
		t.Fatalf("expected token in output: %q", out)
	}
}

func TestFormatScopes(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if got := FormatScopes(nil); got != "<none>" {
			t.Fatalf("want <none>, got %q", got)
		}
	})
	t.Run("single", func(t *testing.T) {
		if got := FormatScopes([]string{"read"}); got != "read" {
			t.Fatalf("want %q, got %q", "read", got)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		got := FormatScopes([]string{"read", "write", "admin"})
		if got != "read, write, admin" {
			t.Fatalf("want %q, got %q", "read, write, admin", got)
		}
	})
}

// ---------------------------------------------------------------------------
// table.go
// ---------------------------------------------------------------------------

func TestFormatAge(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := FormatAge(nil); got != "<unknown>" {
			t.Fatalf("want <unknown>, got %q", got)
		}
	})
	t.Run("seconds", func(t *testing.T) {
		ts := timestamppb.New(time.Now().Add(-30 * time.Second))
		got := FormatAge(ts)
		if !strings.HasSuffix(got, "s") {
			t.Fatalf("expected seconds suffix, got %q", got)
		}
	})
	t.Run("minutes", func(t *testing.T) {
		ts := timestamppb.New(time.Now().Add(-5 * time.Minute))
		got := FormatAge(ts)
		if !strings.HasSuffix(got, "m") {
			t.Fatalf("expected minutes suffix, got %q", got)
		}
	})
	t.Run("hours", func(t *testing.T) {
		ts := timestamppb.New(time.Now().Add(-3 * time.Hour))
		got := FormatAge(ts)
		if !strings.HasSuffix(got, "h") {
			t.Fatalf("expected hours suffix, got %q", got)
		}
	})
	t.Run("days", func(t *testing.T) {
		ts := timestamppb.New(time.Now().Add(-72 * time.Hour))
		got := FormatAge(ts)
		if !strings.HasSuffix(got, "d") {
			t.Fatalf("expected days suffix, got %q", got)
		}
	})
}

func TestFormatTimestamp(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := FormatTimestamp(nil); got != "<none>" {
			t.Fatalf("want <none>, got %q", got)
		}
	})
	t.Run("valid", func(t *testing.T) {
		now := time.Now().UTC()
		ts := timestamppb.New(now)
		got := FormatTimestamp(ts)
		want := now.Format(time.RFC3339)
		if got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})
}

func TestFormatLabels(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := FormatLabels(nil); got != "<none>" {
			t.Fatalf("want <none>, got %q", got)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if got := FormatLabels(map[string]string{}); got != "<none>" {
			t.Fatalf("want <none>, got %q", got)
		}
	})
	t.Run("single", func(t *testing.T) {
		got := FormatLabels(map[string]string{"env": "prod"})
		if got != "env=prod" {
			t.Fatalf("want %q, got %q", "env=prod", got)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		got := FormatLabels(map[string]string{"a": "1", "b": "2"})
		// Map iteration order is non-deterministic; check both labels are present.
		if !strings.Contains(got, "a=1") || !strings.Contains(got, "b=2") {
			t.Fatalf("expected both labels, got %q", got)
		}
		if !strings.Contains(got, ",") {
			t.Fatalf("expected comma separator, got %q", got)
		}
	})
}

func TestFormatEnum(t *testing.T) {
	tests := []struct {
		name   string
		enum   string
		prefix string
		want   string
	}{
		{"healthy", "CLUSTER_HEALTH_STATUS_HEALTHY", "CLUSTER_HEALTH_STATUS_", "Healthy"},
		{"degraded", "CLUSTER_HEALTH_STATUS_DEGRADED", "CLUSTER_HEALTH_STATUS_", "Degraded"},
		{"unspecified", "CLUSTER_HEALTH_STATUS_UNSPECIFIED", "CLUSTER_HEALTH_STATUS_", "Unknown"},
		{"empty after prefix", "SOME_PREFIX_", "SOME_PREFIX_", "Unknown"},
		{"no prefix match", "HEALTHY", "NONEXISTENT_", "Healthy"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatEnum(tc.enum, tc.prefix)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "0s"},
		{"seconds", 45 * time.Second, "45s"},
		{"minutes", 5 * time.Minute, "5m"},
		{"hours", 3 * time.Hour, "3h"},
		{"days", 48 * time.Hour, "2d"},
		{"just under minute", 59 * time.Second, "59s"},
		{"just under hour", 59 * time.Minute, "59m"},
		{"just under day", 23 * time.Hour, "23h"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatDuration(tc.duration)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}
