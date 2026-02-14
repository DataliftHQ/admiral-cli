package cmd

import (
	"bytes"
	"testing"

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
