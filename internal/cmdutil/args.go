package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ExactArgs returns a PositionalArgs validator that requires exactly n arguments.
// When the count is wrong it prints the command's help followed by the error.
func ExactArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			_ = cmd.Help()
			_, _ = fmt.Fprintln(cmd.ErrOrStderr())
			return fmt.Errorf("requires %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// MinimumNArgs returns a PositionalArgs validator that requires at least n arguments.
// When the count is wrong it prints the command's help followed by the error.
func MinimumNArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			_ = cmd.Help()
			_, _ = fmt.Fprintln(cmd.ErrOrStderr())
			return fmt.Errorf("requires at least %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// RangeArgs returns a PositionalArgs validator that requires between min and max arguments.
// When the count is wrong it prints the command's help followed by the error.
func RangeArgs(min, max int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < min || len(args) > max {
			_ = cmd.Help()
			_, _ = fmt.Fprintln(cmd.ErrOrStderr())
			return fmt.Errorf("accepts between %d and %d arg(s), received %d", min, max, len(args))
		}
		return nil
	}
}
