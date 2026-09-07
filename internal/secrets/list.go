// Package secrets discovers YAML/JSON files under a directory tree and
// determines whether each one is SOPS-encrypted.
package secrets

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Status is the SOPS encryption status of a scanned file.
type Status int

const (
	// StatusUnknown means the file's status could not be determined,
	// typically because it failed to parse.
	StatusUnknown Status = iota
	StatusEncrypted
	StatusPlaintext
)

func (s Status) String() string {
	switch s {
	case StatusEncrypted:
		return "Encrypted"
	case StatusPlaintext:
		return "Plaintext"
	default:
		return "Unknown"
	}
}

// FileEntry describes one discovered file and its encryption status.
type FileEntry struct {
	// Path is the file's path relative to the scan root.
	Path string
	// Status is the file's SOPS encryption status.
	Status Status
	// Err is set when Status is StatusUnknown because the file could not
	// be parsed.
	Err error
}

// sopsConfigFile is excluded from scan results: it matches the yaml
// extension filter but is SOPS's own config, not a secret file.
const sopsConfigFile = ".sops.yaml"

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

// List recursively scans root for .yaml/.yml/.json files and reports each
// file's SOPS encryption status. When root is inside a git work tree,
// gitignored paths (e.g. vendor/, node_modules/, build caches) are
// excluded the same way `git status` would exclude them; otherwise it
// falls back to skipping .git, node_modules, and vendor directories at
// any depth.
func List(root string) ([]FileEntry, error) {
	rels, ok := gitTrackedFiles(root)
	if !ok {
		var err error
		rels, err = walkFiles(root)
		if err != nil {
			return nil, err
		}
	}

	var entries []FileEntry
	for _, rel := range rels {
		if filepath.Base(rel) == sopsConfigFile || !isYAMLOrJSON(rel) {
			continue
		}

		status, statusErr := detectStatus(filepath.Join(root, rel))
		entries = append(entries, FileEntry{
			Path:   rel,
			Status: status,
			Err:    statusErr,
		})
	}

	return entries, nil
}

// gitTrackedFiles lists every file under root that git would not ignore
// (tracked files plus untracked-but-not-ignored ones), relative to root.
// ok is false when root is not inside a git work tree or git is
// unavailable, signaling the caller to fall back to a plain directory
// walk.
func gitTrackedFiles(root string) (rels []string, ok bool) {
	out, err := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil, false
	}

	for rel := range strings.SplitSeq(string(out), "\x00") {
		if rel != "" {
			rels = append(rels, rel)
		}
	}
	return rels, true
}

// walkFiles lists every file under root, skipping .git, node_modules, and
// vendor directories at any depth, relative to root.
func walkFiles(root string) ([]string, error) {
	var rels []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path != root && ignoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rels = append(rels, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return rels, nil
}

func isYAMLOrJSON(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml", ".json":
		return true
	default:
		return false
	}
}

// detectStatus reports whether the file at path carries SOPS metadata,
// by checking for a top-level "sops" key. It does not shell out to sops.
func detectStatus(path string) (Status, error) {
	doc, err := readTopLevelKeys(path)
	if err != nil {
		return StatusUnknown, err
	}

	if _, ok := doc["sops"]; ok {
		return StatusEncrypted, nil
	}
	return StatusPlaintext, nil
}

func readTopLevelKeys(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	doc := map[string]any{}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
	default:
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, err
		}
	}

	return doc, nil
}
