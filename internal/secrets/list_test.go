package secrets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const encryptedYAML = `data: ENC[AES256_GCM,data:...,type:str]
sops:
    kms: []
    version: 3.8.1
`

const plaintextYAML = `data: hello
`

const encryptedJSON = `{"data":"ENC[...]","sops":{"version":"3.8.1"}}`

const plaintextJSON = `{"data":"hello"}`

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestList_FiltersToYAMLAndJSON(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "secrets.yaml"), plaintextYAML)
	writeFile(t, filepath.Join(root, "config.json"), plaintextJSON)
	writeFile(t, filepath.Join(root, "README.md"), "not a secret file")
	writeFile(t, filepath.Join(root, "notes.txt"), "not a secret file")

	entries, err := List(root)
	require.NoError(t, err)

	var paths []string
	for _, e := range entries {
		paths = append(paths, e.Path)
	}
	require.ElementsMatch(t, []string{"secrets.yaml", "config.json"}, paths)
}

func TestList_ExcludesIgnoredDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "secrets.yaml"), plaintextYAML)
	writeFile(t, filepath.Join(root, ".git", "config.yaml"), plaintextYAML)
	writeFile(t, filepath.Join(root, "node_modules", "pkg", "data.json"), plaintextJSON)
	writeFile(t, filepath.Join(root, "vendor", "lib", "data.yaml"), plaintextYAML)
	writeFile(t, filepath.Join(root, "deep", "nested", "vendor", "data.yaml"), plaintextYAML)

	entries, err := List(root)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "secrets.yaml", entries[0].Path)
}

func TestList_DetectsEncryptedAndPlaintextStatus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "enc.yaml"), encryptedYAML)
	writeFile(t, filepath.Join(root, "plain.yaml"), plaintextYAML)
	writeFile(t, filepath.Join(root, "enc.json"), encryptedJSON)
	writeFile(t, filepath.Join(root, "plain.json"), plaintextJSON)

	entries, err := List(root)
	require.NoError(t, err)

	statusByPath := map[string]Status{}
	for _, e := range entries {
		statusByPath[e.Path] = e.Status
	}

	require.Equal(t, StatusEncrypted, statusByPath["enc.yaml"])
	require.Equal(t, StatusPlaintext, statusByPath["plain.yaml"])
	require.Equal(t, StatusEncrypted, statusByPath["enc.json"])
	require.Equal(t, StatusPlaintext, statusByPath["plain.json"])
}

func TestList_MalformedFileDoesNotAbortScan(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "broken.yaml"), "{ this is not: valid: yaml: [")
	writeFile(t, filepath.Join(root, "plain.yaml"), plaintextYAML)

	entries, err := List(root)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	statusByPath := map[string]FileEntry{}
	for _, e := range entries {
		statusByPath[e.Path] = e
	}

	require.Equal(t, StatusUnknown, statusByPath["broken.yaml"].Status)
	require.Error(t, statusByPath["broken.yaml"].Err)
	require.Equal(t, StatusPlaintext, statusByPath["plain.yaml"].Status)
	require.NoError(t, statusByPath["plain.yaml"].Err)
}
