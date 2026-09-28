// Package project describes the directory Shipcheck is inspecting and
// provides read-only helpers for looking at its files and Git state.
package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Project is a directory on disk that Shipcheck inspects.
// Every helper is read-only: Shipcheck never modifies project files.
type Project struct {
	Root string
	Git  *Git
}

// Open resolves dir to an absolute path and verifies it is a directory.
func Open(dir string) (*Project, error) {
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", dir, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("directory %q does not exist", dir)
		}
		return nil, fmt.Errorf("open %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", dir)
	}
	return &Project{Root: abs, Git: NewGit(abs)}, nil
}

// Path joins rel onto the project root.
func (p *Project) Path(rel string) string {
	return filepath.Join(p.Root, filepath.FromSlash(rel))
}

// Exists reports whether rel exists (file or directory).
func (p *Project) Exists(rel string) bool {
	_, err := os.Stat(p.Path(rel))
	return err == nil
}

// IsFile reports whether rel exists and is a regular file.
func (p *Project) IsFile(rel string) bool {
	info, err := os.Stat(p.Path(rel))
	return err == nil && info.Mode().IsRegular()
}

// IsDir reports whether rel exists and is a directory.
func (p *Project) IsDir(rel string) bool {
	info, err := os.Stat(p.Path(rel))
	return err == nil && info.IsDir()
}

// ReadFile reads rel. Files larger than maxRead are rejected so a stray
// binary or log file can never make Shipcheck slow.
func (p *Project) ReadFile(rel string) ([]byte, error) {
	const maxRead = 4 << 20
	path := p.Path(rel)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxRead {
		return nil, fmt.Errorf("%s is larger than %d bytes", rel, maxRead)
	}
	return os.ReadFile(path)
}

// FindFile returns the first entry in the project root whose name matches
// one of candidates, ignoring case. Candidates are tried in order, so the
// preferred spelling should come first. It returns "" when nothing matches.
func (p *Project) FindFile(candidates ...string) string {
	entries, err := os.ReadDir(p.Root)
	if err != nil {
		return ""
	}
	for _, want := range candidates {
		for _, e := range entries {
			if e.Type().IsRegular() && strings.EqualFold(e.Name(), want) {
				return e.Name()
			}
		}
	}
	return ""
}

// HasFileWithExt reports whether dir (relative to the root) directly
// contains a regular file with one of the given extensions.
func (p *Project) HasFileWithExt(dir string, exts ...string) bool {
	entries, err := os.ReadDir(p.Path(dir))
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		for _, want := range exts {
			if ext == want {
				return true
			}
		}
	}
	return false
}
