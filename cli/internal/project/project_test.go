package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir string, files map[string]string) {
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

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{"file.txt": "x"})

	p, err := Open(dir)
	if err != nil || !filepath.IsAbs(p.Root) {
		t.Fatalf("Open = %+v, %v", p, err)
	}
	if _, err := Open(filepath.Join(dir, "missing")); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("missing dir error = %v", err)
	}
	if _, err := Open(filepath.Join(dir, "file.txt")); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("file error = %v", err)
	}
}

func TestFileHelpers(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"readme.MD":        "hi",
		"docs/guide.md":    "",
		"tests/test_x.py":  "",
		"tests/fixture.js": "",
	})
	p, _ := Open(dir)

	if !p.IsFile("readme.MD") || p.IsFile("docs") || !p.IsDir("docs") || !p.Exists("docs/guide.md") {
		t.Error("IsFile/IsDir/Exists disagree with the file system")
	}
	if got := p.FindFile("README.md", "README"); got != "readme.MD" {
		t.Errorf("FindFile case-insensitive = %q", got)
	}
	if got := p.FindFile("docs"); got != "" {
		t.Errorf("FindFile must ignore directories, got %q", got)
	}
	if !p.HasFileWithExt("tests", ".py") || p.HasFileWithExt("tests", ".go") || p.HasFileWithExt("missing", ".py") {
		t.Error("HasFileWithExt")
	}
	data, err := p.ReadFile("readme.MD")
	if err != nil || string(data) != "hi" {
		t.Errorf("ReadFile = %q, %v", data, err)
	}
}

func TestIsRepoFallbackWithoutGit(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{".git/HEAD": "ref: refs/heads/main\n", "sub/x": ""})
	t.Setenv("PATH", "")

	sub, _ := Open(filepath.Join(dir, "sub"))
	if sub.Git.Available() {
		t.Skip("git still found without PATH")
	}
	if !sub.Git.IsRepo() {
		t.Error("expected .git in a parent directory to count as a repository")
	}
	if _, err := sub.Git.Changes(); err != ErrGitNotInstalled {
		t.Errorf("Changes error = %v", err)
	}
}
