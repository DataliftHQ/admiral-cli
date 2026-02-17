package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"go.admiral.io/cli/internal/version"
)

var testversion = version.Version{
	GitVersion: "1.2.3",
}

func TestRootCmd(t *testing.T) {
	mem := &exitMemento{}
	Execute(testversion, mem.Exit, []string{"-h"})
	require.Equal(t, 0, mem.code)
}

func TestRootCmdHelp(t *testing.T) {
	mem := &exitMemento{}
	cmd := newRootCmd(testversion, mem.Exit).cmd
	cmd.SetArgs([]string{"-h"})
	require.NoError(t, cmd.Execute())
	require.Equal(t, 0, mem.code)
}

func TestRootCmdVersion(t *testing.T) {
	var b bytes.Buffer
	mem := &exitMemento{}
	cmd := newRootCmd(testversion, mem.Exit).cmd
	cmd.SetOut(&b)
	cmd.SetArgs([]string{"--version"})
	require.NoError(t, cmd.Execute())
	require.Equal(t, "GitVersion:  1.2.3\nGitCommit:   \nBuildDate:   \nBuiltBy:     \nGoVersion:   \nCompiler:    \nPlatform:    \n", b.String())
	require.Equal(t, 0, mem.code)
}

func TestRootCmdExitCode(t *testing.T) {
	setup(t)
	mem := &exitMemento{}
	cmd := newRootCmd(testversion, mem.Exit)
	cmd.Execute([]string{})
	require.Equal(t, 0, mem.code)
}

// ---------------------------------------------------------------------------
// Command tree structure
// ---------------------------------------------------------------------------

func TestRootCmd_HasExpectedSubcommands(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	expected := []string{
		"auth", "cluster", "version", "completion", "whoami", "use",
	}

	names := make([]string, 0, len(root.Commands()))
	for _, c := range root.Commands() {
		names = append(names, c.Name())
	}

	for _, want := range expected {
		require.Contains(t, names, want, "missing subcommand %q", want)
	}
}

func TestRootCmd_AuthSubcommands(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	auth, _, err := root.Find([]string{"auth"})
	require.NoError(t, err)

	expected := []string{"login", "logout"}
	names := subcmdNames(auth)
	for _, want := range expected {
		require.Contains(t, names, want, "auth missing subcommand %q", want)
	}
}

func TestRootCmd_ClusterSubcommands(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	cluster, _, err := root.Find([]string{"cluster"})
	require.NoError(t, err)

	expected := []string{"list", "get", "create", "update", "delete", "status", "token"}
	names := subcmdNames(cluster)
	for _, want := range expected {
		require.Contains(t, names, want, "cluster missing subcommand %q", want)
	}
}

func TestRootCmd_ClusterTokenSubcommands(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	tokenCmd, _, err := root.Find([]string{"cluster", "token"})
	require.NoError(t, err)

	expected := []string{"create", "list", "get", "revoke"}
	names := subcmdNames(tokenCmd)
	for _, want := range expected {
		require.Contains(t, names, want, "cluster token missing subcommand %q", want)
	}
}

// ---------------------------------------------------------------------------
// Persistent flags
// ---------------------------------------------------------------------------

func TestRootCmd_PersistentFlags(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	flags := []struct {
		name     string
		shortcut string
	}{
		{"server", "s"},
		{"output", "o"},
		{"verbose", "v"},
		{"config-dir", ""},
		{"plaintext", ""},
		{"insecure", "i"},
	}

	for _, f := range flags {
		t.Run(f.name, func(t *testing.T) {
			flag := root.PersistentFlags().Lookup(f.name)
			require.NotNil(t, flag, "persistent flag %q not found", f.name)
			if f.shortcut != "" {
				require.Equal(t, f.shortcut, flag.Shorthand)
			}
		})
	}
}

func TestRootCmd_HiddenFlags(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	for _, name := range []string{"issuer", "client-id", "scopes"} {
		flag := root.PersistentFlags().Lookup(name)
		require.NotNil(t, flag, "hidden flag %q not found", name)
		require.True(t, flag.Hidden, "flag %q should be hidden", name)
	}
}

func TestRootCmd_FlagDefaults(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd

	tests := []struct {
		name     string
		defValue string
	}{
		{"server", "api.admiral.io:443"},
		{"output", "table"},
		{"issuer", "https://auth.admiral.io"},
		{"plaintext", "false"},
		{"insecure", "false"},
		{"verbose", "false"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flag := root.PersistentFlags().Lookup(tc.name)
			require.NotNil(t, flag)
			require.Equal(t, tc.defValue, flag.DefValue)
		})
	}
}

// ---------------------------------------------------------------------------
// Output format validation
// ---------------------------------------------------------------------------

func TestRootCmd_InvalidOutputFormat(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"--output", "xml", "version"})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid output format")
}

func TestRootCmd_ValidOutputFormats(t *testing.T) {
	for _, format := range []string{"table", "json", "yaml", "wide"} {
		t.Run(format, func(t *testing.T) {
			var buf bytes.Buffer
			mem := &exitMemento{}
			root := newRootCmd(testversion, mem.Exit).cmd
			root.SetOut(&buf)
			root.SetArgs([]string{"--output", format, "version"})

			err := root.Execute()
			require.NoError(t, err)
		})
	}
}

// ---------------------------------------------------------------------------
// Error handling
// ---------------------------------------------------------------------------

func TestExitError_Error(t *testing.T) {
	eerr := &exitError{
		err:  &simpleError{msg: "something failed"},
		code: 2,
	}
	require.Equal(t, "something failed", eerr.Error())
	require.Equal(t, 2, eerr.code)
}

func TestExitError_DefaultCode(t *testing.T) {
	eerr := &exitError{
		err:  &simpleError{msg: "generic error"},
		code: 1,
	}
	require.Equal(t, 1, eerr.code)
	require.Equal(t, "generic error", eerr.Error())
}

func TestRootCmd_ExecuteWithError(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit)

	// Invalid output format triggers a PersistentPreRunE error
	root.Execute([]string{"--output", "xml", "version"})

	require.Equal(t, 1, mem.code)
}

func TestRootCmd_ExecuteWithExitError(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit)

	// Missing required flag triggers an error that flows through Execute
	root.Execute([]string{"cluster", "create"})
	require.Equal(t, 1, mem.code)
}

func TestRootCmd_ExecuteWithExitErrorCode(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit)

	// Inject a command that returns an exitError with a custom code
	root.cmd.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(cmd *cobra.Command, args []string) error {
			return &exitError{err: &simpleError{msg: "custom exit"}, code: 42}
		},
	})

	root.Execute([]string{"fail"})
	require.Equal(t, 42, mem.code)
}

// ---------------------------------------------------------------------------
// Version command
// ---------------------------------------------------------------------------

func TestVersionCmd(t *testing.T) {
	var buf bytes.Buffer
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetOut(&buf)
	root.SetArgs([]string{"version"})

	require.NoError(t, root.Execute())

	out := buf.String()
	require.Contains(t, out, "GitVersion:")
	require.Contains(t, out, "1.2.3")
	require.Contains(t, out, "GoVersion:")
	require.Contains(t, out, "Platform:")
}

func TestVersionCmd_NoArgs(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"version", "extra"})

	err := root.Execute()
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Completion command
// ---------------------------------------------------------------------------

func TestCompletionCmd_AllShells(t *testing.T) {
	// Completion commands write to os.Stdout directly (not cmd.OutOrStdout()),
	// so we only verify execution succeeds without error.
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			mem := &exitMemento{}
			root := newRootCmd(testversion, mem.Exit).cmd
			root.SetArgs([]string{"completion", shell})

			require.NoError(t, root.Execute())
		})
	}
}

func TestCompletionCmd_InvalidShell(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"completion", "tcsh"})

	err := root.Execute()
	require.Error(t, err)
}

func TestCompletionCmd_NoArgs(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	require.Error(t, err)
}

func TestCompletionCmd_TooManyArgs(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"completion", "bash", "extra"})

	err := root.Execute()
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Help output content
// ---------------------------------------------------------------------------

func TestRootCmd_HelpContainsDescription(t *testing.T) {
	var buf bytes.Buffer
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetOut(&buf)
	root.SetArgs([]string{"-h"})

	require.NoError(t, root.Execute())

	out := buf.String()
	require.Contains(t, out, "Admiral")
	require.Contains(t, out, "Platform Orchestrator")
}

func TestSubcommand_HelpOutput(t *testing.T) {
	tests := []struct {
		args     []string
		contains string
	}{
		{[]string{"auth", "-h"}, "authentication"},
		{[]string{"cluster", "-h"}, "clusters"},
		{[]string{"completion", "-h"}, "completion"},
	}

	for _, tc := range tests {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			var buf bytes.Buffer
			mem := &exitMemento{}
			root := newRootCmd(testversion, mem.Exit).cmd
			root.SetOut(&buf)
			root.SetArgs(tc.args)

			require.NoError(t, root.Execute())
			require.Contains(t, strings.ToLower(buf.String()), tc.contains)
		})
	}
}

// ---------------------------------------------------------------------------
// Arg validation on leaf commands
// ---------------------------------------------------------------------------

func TestLeafCommand_ArgValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"cluster get needs 1 arg", []string{"cluster", "get"}},
		{"cluster get rejects 2 args", []string{"cluster", "get", "a", "b"}},
		{"cluster delete needs 1 arg", []string{"cluster", "delete"}},
		{"cluster update needs 1 arg", []string{"cluster", "update"}},
		{"cluster status needs 1 arg", []string{"cluster", "status"}},
		{"cluster token get needs 2 args", []string{"cluster", "token", "get"}},
		{"cluster token get rejects 3 args", []string{"cluster", "token", "get", "a", "b", "c"}},
		{"cluster token revoke needs 2 args", []string{"cluster", "token", "revoke"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mem := &exitMemento{}
			root := newRootCmd(testversion, mem.Exit).cmd
			root.SetArgs(tc.args)

			err := root.Execute()
			require.Error(t, err)
		})
	}
}

// Leaf commands that take no args should reject stray args.
func TestLeafCommand_NoArgsValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"cluster list rejects args", []string{"cluster", "list", "extra"}},
		{"cluster create rejects args", []string{"cluster", "create", "extra"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mem := &exitMemento{}
			root := newRootCmd(testversion, mem.Exit).cmd
			root.SetArgs(tc.args)

			err := root.Execute()
			require.Error(t, err)
		})
	}
}

// ---------------------------------------------------------------------------
// Required flags
// ---------------------------------------------------------------------------

func TestClusterCreate_RequiresNameFlag(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"cluster", "create"})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "name")
}

func TestClusterTokenCreate_RequiresNameFlag(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"cluster", "token", "create", "cluster-id"})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "name")
}

// ---------------------------------------------------------------------------
// Verbose mode
// ---------------------------------------------------------------------------

func TestRootCmd_VerboseFlag(t *testing.T) {
	var buf bytes.Buffer
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetOut(&buf)
	root.SetArgs([]string{"-v", "version"})

	require.NoError(t, root.Execute())
}

// ---------------------------------------------------------------------------
// Use command
// ---------------------------------------------------------------------------

func TestUseCmd_SetApp(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetOut(&buf)
	root.SetArgs([]string{"--config-dir", dir, "use", "my-app"})

	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "my-app")
}

func TestUseCmd_ShowContext(t *testing.T) {
	dir := t.TempDir()
	mem := &exitMemento{}

	// Set context first.
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"--config-dir", dir, "use", "my-app"})
	require.NoError(t, root.Execute())

	// Show context.
	var buf bytes.Buffer
	root2 := newRootCmd(testversion, mem.Exit).cmd
	root2.SetOut(&buf)
	root2.SetArgs([]string{"--config-dir", dir, "use"})
	require.NoError(t, root2.Execute())

	require.Contains(t, buf.String(), "my-app")
}

func TestUseCmd_ShowNoContext(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetOut(&buf)
	root.SetArgs([]string{"--config-dir", dir, "use"})

	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "No active app context")
}

func TestUseCmd_Clear(t *testing.T) {
	dir := t.TempDir()
	mem := &exitMemento{}

	// Set context first.
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"--config-dir", dir, "use", "my-app"})
	require.NoError(t, root.Execute())

	// Clear it.
	var buf bytes.Buffer
	root2 := newRootCmd(testversion, mem.Exit).cmd
	root2.SetOut(&buf)
	root2.SetArgs([]string{"--config-dir", dir, "use", "--clear"})
	require.NoError(t, root2.Execute())
	require.Contains(t, buf.String(), "Context cleared")

	// Verify it's gone.
	var buf2 bytes.Buffer
	root3 := newRootCmd(testversion, mem.Exit).cmd
	root3.SetOut(&buf2)
	root3.SetArgs([]string{"--config-dir", dir, "use"})
	require.NoError(t, root3.Execute())
	require.Contains(t, buf2.String(), "No active app context")
}

func TestUseCmd_RejectsTooManyArgs(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"use", "app1", "app2"})

	err := root.Execute()
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Whoami command
// ---------------------------------------------------------------------------

func TestWhoamiCmd_RejectsArgs(t *testing.T) {
	mem := &exitMemento{}
	root := newRootCmd(testversion, mem.Exit).cmd
	root.SetArgs([]string{"whoami", "extra"})

	err := root.Execute()
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func subcmdNames(cmd *cobra.Command) []string {
	names := make([]string, 0, len(cmd.Commands()))
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	return names
}

type simpleError struct {
	msg string
}

func (e *simpleError) Error() string {
	return e.msg
}
