package secrets

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const sopsBinary = "sops"

// CheckAvailable reports whether the sops binary is resolvable on PATH.
func CheckAvailable() error {
	if _, err := exec.LookPath(sopsBinary); err != nil {
		return fmt.Errorf("sops binary not found on PATH: %w", err)
	}
	return nil
}

// View returns the decrypted content of the file at path. The file on
// disk is left unchanged.
func View(path string) (string, error) {
	return runSopsOnFile(path, "-d")
}

// Decrypt decrypts the file at path in place.
func Decrypt(path string) error {
	_, err := runSopsOnFile(path, "-d", "-i")
	return err
}

// Encrypt encrypts the file at path in place, using the nearest
// .sops.yaml's creation rules to select keys.
func Encrypt(path string) error {
	_, err := runSopsOnFile(path, "-e", "-i")
	return err
}

// Edit runs `sops <path>` as a foreground subprocess with inherited
// stdio: sops decrypts to a temp file, launches $EDITOR on it, and
// re-encrypts on save. The caller is responsible for suspending its own
// terminal UI before calling Edit.
func Edit(path string) error {
	cmd := fileCommand(path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runSopsOnFile runs sops with the given flags followed by path, from
// within path's own directory. sops's .sops.yaml discovery walks up from
// the process's working directory, so this keeps that discovery correct
// regardless of where sops-tui itself was launched from.
func runSopsOnFile(path string, flags ...string) (string, error) {
	cmd := fileCommand(path, flags...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("sops %s %s: %s", strings.Join(flags, " "), filepath.Base(path), msg)
	}

	return stdout.String(), nil
}

func fileCommand(path string, flagsBeforeFile ...string) *exec.Cmd {
	args := append(append([]string{}, flagsBeforeFile...), filepath.Base(path))
	cmd := exec.Command(sopsBinary, args...)
	cmd.Dir = filepath.Dir(path)
	return cmd
}
