package root

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePath_EmptyDefaultsToCwd(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	got, err := ResolvePath("")
	require.NoError(t, err)
	require.Equal(t, cwd, got)
}

func TestResolvePath_RelativeResolvedToAbsolute(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "secrets")
	require.NoError(t, os.Mkdir(sub, 0o755))

	t.Chdir(dir)

	got, err := ResolvePath("secrets")
	require.NoError(t, err)
	require.Equal(t, sub, got)
	require.True(t, filepath.IsAbs(got))
}

func TestResolvePath_AbsoluteUnchanged(t *testing.T) {
	dir := t.TempDir()

	got, err := ResolvePath(dir)
	require.NoError(t, err)
	require.Equal(t, dir, got)
}
