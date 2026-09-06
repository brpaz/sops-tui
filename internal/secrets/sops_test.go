package secrets

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ageFixture generates a throwaway age keypair, points sops at its
// private key via SOPS_AGE_KEY_FILE, and returns the public key for use
// in a test-local .sops.yaml.
func ageFixture(t *testing.T) (publicKey string) {
	t.Helper()

	out, err := exec.Command("age-keygen").Output()
	require.NoError(t, err, "age-keygen must be installed to run these tests")

	keyFile := filepath.Join(t.TempDir(), "keys.txt")
	require.NoError(t, os.WriteFile(keyFile, out, 0o600))
	t.Setenv("SOPS_AGE_KEY_FILE", keyFile)

	for _, line := range strings.Split(string(out), "\n") {
		if pk, ok := strings.CutPrefix(line, "# public key: "); ok {
			return pk
		}
	}
	t.Fatal("could not find public key in age-keygen output")
	return ""
}

func writeSopsConfig(t *testing.T, dir, pathRegex, publicKey string) {
	t.Helper()
	content := "creation_rules:\n  - path_regex: '" + pathRegex + "'\n    age: '" + publicKey + "'\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".sops.yaml"), []byte(content), 0o644))
}

func TestCheckAvailable(t *testing.T) {
	require.NoError(t, CheckAvailable())
}

func TestEncryptThenDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, dir, ".*\\.yaml$", publicKey)

	path := filepath.Join(dir, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	require.NoError(t, Encrypt(path))

	status, err := detectStatus(path)
	require.NoError(t, err)
	require.Equal(t, StatusEncrypted, status)

	decrypted, err := View(path)
	require.NoError(t, err)
	require.Contains(t, decrypted, "password: hunter2")

	// View must not touch disk.
	statusAfterView, err := detectStatus(path)
	require.NoError(t, err)
	require.Equal(t, StatusEncrypted, statusAfterView)

	require.NoError(t, Decrypt(path))

	statusAfterDecrypt, err := detectStatus(path)
	require.NoError(t, err)
	require.Equal(t, StatusPlaintext, statusAfterDecrypt)

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), "password: hunter2")
}

func TestEncrypt_NoMatchingCreationRule(t *testing.T) {
	dir := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, dir, "^matched\\.yaml$", publicKey)

	path := filepath.Join(dir, "nomatch.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	err := Encrypt(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nomatch.yaml")

	status, statusErr := detectStatus(path)
	require.NoError(t, statusErr)
	require.Equal(t, StatusPlaintext, status)
}

func TestEdit_RoundTripsThroughEditor(t *testing.T) {
	dir := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, dir, ".*\\.yaml$", publicKey)

	path := filepath.Join(dir, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, Encrypt(path))

	editorScript := filepath.Join(dir, "stub-editor.sh")
	require.NoError(t, os.WriteFile(editorScript, []byte(
		"#!/bin/sh\necho 'appended: true' >> \"$1\"\n",
	), 0o755))
	t.Setenv("EDITOR", editorScript)

	require.NoError(t, Edit(path))

	status, err := detectStatus(path)
	require.NoError(t, err)
	require.Equal(t, StatusEncrypted, status, "sops re-encrypts on save")

	decrypted, err := View(path)
	require.NoError(t, err)
	require.Contains(t, decrypted, "password: hunter2")
	require.Contains(t, decrypted, "appended: true")
}
