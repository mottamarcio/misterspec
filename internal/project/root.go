package project

import (
	"errors"
	"os"
	"path/filepath"
)

// Project represents the root of a misterspec-managed repository, once
// detected (FR-001, FR-002).
type Project struct {
	// Root is the project's root directory, as an absolute, normalized
	// path — identical regardless of which nested subdirectory Detect
	// was called from.
	Root string
	// Config is this project's fully resolved configuration.
	Config Configuration
}

// ErrNotInitialized is returned by Detect when startDir is not inside any
// misterspec project — i.e. no .misterspec/config.yaml was found anywhere
// from startDir up to the filesystem root (FR-003). It is always a
// distinct condition from ErrInvalidConfiguration: this means "no project
// was found here," never "a project was found but is broken."
var ErrNotInitialized = errors.New("project: not initialized")

// Detect walks upward from startDir, looking for the first ancestor
// directory containing .misterspec/config.yaml. When found, it loads and
// validates that configuration and returns the resulting Project.
//
// Detect never returns a partially valid Project: it returns exactly one
// of a populated *Project, ErrNotInitialized (no project found in
// startDir's ancestry), or an error wrapping ErrInvalidConfiguration (a
// project was found, but its configuration is broken) — see Load.
func Detect(startDir string) (*Project, error) {
	root, err := findProjectRoot(startDir)
	if err != nil {
		return nil, err
	}

	cfg, err := Load(root)
	if err != nil {
		// A project marker was found, but its configuration is
		// invalid — this must never be reported as ErrNotInitialized.
		return nil, err
	}

	return &Project{Root: root, Config: cfg}, nil
}

// findProjectRoot walks upward from startDir looking for a directory
// containing .misterspec/config.yaml, returning that directory as an
// absolute, normalized path. It returns ErrNotInitialized if no such
// directory is found before reaching the filesystem root.
func findProjectRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir = filepath.Clean(dir)

	for {
		marker := filepath.Join(dir, configFilePath)
		if info, err := os.Stat(marker); err == nil && !info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root without finding a marker.
			return "", ErrNotInitialized
		}
		dir = parent
	}
}
