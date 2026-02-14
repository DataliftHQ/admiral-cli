package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type exitMemento struct {
	code int
}

func (e *exitMemento) Exit(i int) {
	e.code = i
}

func setup(tb testing.TB) string {
	tb.Helper()

	previous, err := os.Getwd()
	require.NoError(tb, err)

	tb.Cleanup(func() {
		require.NoError(tb, os.Chdir(previous))
	})

	folder := tb.TempDir()
	require.NoError(tb, os.Chdir(folder))

	return folder
}
