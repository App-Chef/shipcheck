// Package testutil builds throwaway projects for tests.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Dir creates a temp directory containing files (path → content).
func Dir(t testing.TB, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	Write(t, dir, files)
	return dir
}

// Write creates files under dir, making parent directories as needed.
func Write(t testing.TB, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// RequireGit skips the test when git is not installed.
func RequireGit(t testing.TB) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

// Git runs a git command in dir with a fixed identity and fails the test
// on error. It returns trimmed output.
func Git(t testing.TB, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Shipcheck Test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Shipcheck Test", "GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_NOSYSTEM=1", "HOME="+dir, "XDG_CONFIG_HOME="+dir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// GitRepo initialises a repository on branch main in dir and commits
// everything in it.
func GitRepo(t testing.TB, dir string) {
	t.Helper()
	RequireGit(t)
	Git(t, dir, "init", "-q")
	Git(t, dir, "symbolic-ref", "HEAD", "refs/heads/main")
	Commit(t, dir, "initial commit")
}

// Commit stages everything and commits it.
func Commit(t testing.TB, dir, msg string) {
	t.Helper()
	Git(t, dir, "add", "-A")
	Git(t, dir, "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", msg)
}
