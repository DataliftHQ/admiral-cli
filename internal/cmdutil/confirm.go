package cmdutil

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ConfirmPrompt asks the user a yes/no question, defaulting to "No".
// It reads from r and writes the prompt to w. Returns true only if
// the user enters "y" or "yes" (case-insensitive).
func ConfirmPrompt(r io.Reader, w io.Writer, question string) (bool, error) {
	Writef(w, "%s [y/N]: ", question)

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("failed to read input: %w", err)
		}
		return false, nil // EOF
	}

	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

// Writef writes formatted output to w, swallowing the return values.
// This is a convenience re-export from the output package to avoid
// circular imports when cmdutil needs to write prompts.
func Writef(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
