package project

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrGitNotInstalled is returned when the git binary is not on PATH.
var ErrGitNotInstalled = errors.New("git is not installed")

// Git runs read-only git commands inside a project.
type Git struct {
	dir string
}

// NewGit returns a Git helper rooted at dir.
func NewGit(dir string) *Git {
	return &Git{dir: dir}
}

// Available reports whether the git binary can be found.
func (g *Git) Available() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// run executes git with args and returns trimmed stdout.
func (g *Git) run(args ...string) (string, error) {
	if !g.Available() {
		return "", ErrGitNotInstalled
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.dir
	// Never prompt for credentials or open an editor/pager.
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_PAGER=cat", "LC_ALL=C")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return "", err
		}
		return "", errors.New(msg)
	}
	return strings.TrimRight(stdout.String(), "\r\n"), nil
}

// IsRepo reports whether the project is inside a Git work tree.
// When git is not installed it falls back to looking for a .git entry
// in the project or any parent directory.
func (g *Git) IsRepo() bool {
	if !g.Available() {
		return findDotGit(g.dir)
	}
	out, err := g.run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// TopLevel returns the root of the work tree.
func (g *Git) TopLevel() (string, error) {
	return g.run("rev-parse", "--show-toplevel")
}

// Changes returns the porcelain status lines for uncommitted changes,
// including untracked files. Paths are relative to the project.
func (g *Git) Changes() ([]string, error) {
	out, err := g.run("status", "--porcelain", "--untracked-files=normal", "--", ".")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// Branch returns the current branch name, or "" when HEAD is detached.
func (g *Git) Branch() (string, error) {
	out, err := g.run("symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		if errors.Is(err, ErrGitNotInstalled) {
			return "", err
		}
		// symbolic-ref exits non-zero with no output on a detached HEAD.
		return "", nil
	}
	return out, nil
}

// IsTracked reports whether rel is tracked by git.
func (g *Git) IsTracked(rel string) bool {
	_, err := g.run("ls-files", "--error-unmatch", "--", rel)
	return err == nil
}

// RevParse resolves a revision to a commit hash.
func (g *Git) RevParse(rev string) (string, error) {
	return g.run("rev-parse", "--verify", "--quiet", rev+"^{commit}")
}

// LatestTag returns the most recent tag reachable from HEAD, or "".
func (g *Git) LatestTag() string {
	out, err := g.run("describe", "--tags", "--abbrev=0")
	if err != nil {
		return ""
	}
	return out
}

func findDotGit(dir string) bool {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}
