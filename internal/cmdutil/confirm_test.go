package cmdutil_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.admiral.io/cli/internal/cmdutil"
)

func TestConfirmPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "y", input: "y\n", expected: true},
		{name: "Y", input: "Y\n", expected: true},
		{name: "yes", input: "yes\n", expected: true},
		{name: "YES", input: "YES\n", expected: true},
		{name: "Yes", input: "Yes\n", expected: true},
		{name: "n", input: "n\n", expected: false},
		{name: "no", input: "no\n", expected: false},
		{name: "empty", input: "\n", expected: false},
		{name: "random", input: "maybe\n", expected: false},
		{name: "y with spaces", input: "  y  \n", expected: true},
		{name: "EOF", input: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			result, err := cmdutil.ConfirmPrompt(strings.NewReader(tt.input), &out, "Continue?")
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
			require.Contains(t, out.String(), "[y/N]")
		})
	}
}
