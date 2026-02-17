package properties

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	props, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, &Properties{}, props)
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	want := &Properties{App: "my-app"}
	require.NoError(t, Save(dir, want))

	got, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")

	require.NoError(t, Save(dir, &Properties{App: "test"}))

	got, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, "test", got.App)
}

func TestSave_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Save(dir, &Properties{App: "test"}))

	info, err := os.Stat(filepath.Join(dir, propertiesFile))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestClear(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, Save(dir, &Properties{App: "my-app"}))
	require.NoError(t, Clear(dir))

	props, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, &Properties{}, props)
}

func TestClear_NoFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Clear(dir))
}

func TestLoad_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, propertiesFile), []byte("not json"), 0600))

	_, err := Load(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse properties")
}
