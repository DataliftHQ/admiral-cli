package env

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveAppForEnv(t *testing.T) {
	t.Run("returns app from context", func(t *testing.T) {
		dir := t.TempDir()
		writeProperties(t, dir, "billing-api")

		app, err := resolveAppForEnv(dir)
		require.NoError(t, err)
		require.Equal(t, "billing-api", app)
	})

	t.Run("errors when no context set", func(t *testing.T) {
		dir := t.TempDir()

		_, err := resolveAppForEnv(dir)
		require.ErrorContains(t, err, "no app context set")
	})

	t.Run("errors when context is empty string", func(t *testing.T) {
		dir := t.TempDir()
		writeProperties(t, dir, "")

		_, err := resolveAppForEnv(dir)
		require.ErrorContains(t, err, "no app context set")
	})
}

func writeProperties(t *testing.T, dir, app string) {
	t.Helper()
	data, err := json.Marshal(map[string]string{"app": app})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "properties.json"), data, 0600))
}
