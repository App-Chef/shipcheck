package checks

import (
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

func TestGitRepo(t *testing.T) {
	testutil.RequireGit(t)
	dir := testutil.Dir(t, map[string]string{"a.txt": "a"})
	expect(t, run(t, GitRepo, newEnv(t, dir, "", nil)), Fail, "Not a Git repository")

	testutil.GitRepo(t, dir)
	expect(t, run(t, GitRepo, newEnv(t, dir, "", nil)), Pass, "Git repository")
}

func TestGitRepoFromSubdirectory(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"web/package.json": "{}"})
	testutil.GitRepo(t, dir)
	expect(t, run(t, GitRepo, newEnv(t, dir+"/web", "", nil)), Pass, "Git repository")
}

func TestGitClean(t *testing.T) {
	testutil.RequireGit(t)
	dir := testutil.Dir(t, map[string]string{"a.txt": "a"})
	expect(t, run(t, GitClean, newEnv(t, dir, "", nil)), Skip, "not a Git repository")

	testutil.GitRepo(t, dir)
	expect(t, run(t, GitClean, newEnv(t, dir, "", nil)), Pass, "Working tree clean")

	testutil.Write(t, dir, map[string]string{"a.txt": "changed"})
	res := run(t, GitClean, newEnv(t, dir, "", nil))
	expect(t, res, Fail, "1 uncommitted change")
	if res.Suggestion == "" || len(res.Details) != 1 {
		t.Errorf("want suggestion and one detail line, got %+v", res)
	}

	testutil.Write(t, dir, map[string]string{"new.txt": "untracked"})
	expect(t, run(t, GitClean, newEnv(t, dir, "", nil)), Fail, "2 uncommitted changes")
}

func TestGitCleanIgnoresGitignored(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{".gitignore": "node_modules/\n"})
	testutil.GitRepo(t, dir)
	testutil.Write(t, dir, map[string]string{"node_modules/x/index.js": ""})
	expect(t, run(t, GitClean, newEnv(t, dir, "", nil)), Pass, "Working tree clean")
}

func TestGitCleanLimitsDetails(t *testing.T) {
	dir := testutil.Dir(t, nil)
	testutil.GitRepo(t, dir)
	files := map[string]string{}
	for _, c := range "abcdefghijklmn" {
		files[string(c)+".txt"] = "x"
	}
	testutil.Write(t, dir, files)
	res := run(t, GitClean, newEnv(t, dir, "", nil))
	expect(t, res, Fail, "14 uncommitted changes")
	if len(res.Details) != 11 || res.Details[10] != "… and 4 more" {
		t.Errorf("details = %q", res.Details)
	}
}

func TestGitBranch(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"a.txt": "a"})
	expect(t, run(t, GitBranch, newEnv(t, dir, "", nil)), Skip, "not a Git repository")

	testutil.GitRepo(t, dir)
	expect(t, run(t, GitBranch, newEnv(t, dir, "", nil)), Pass, "On branch main")

	testutil.Git(t, dir, "checkout", "-q", "-b", "feature/login")
	res := run(t, GitBranch, newEnv(t, dir, "", nil))
	expect(t, res, Warn, "On branch feature/login, not main")

	cfg := "version: 1\ngit:\n  protected_branches: [release, feature/login]\n"
	expect(t, run(t, GitBranch, newEnv(t, dir, cfg, nil)), Pass, "On branch feature/login")

	cfg = "version: 1\ngit:\n  protected_branches: [release, prod]\n"
	expect(t, run(t, GitBranch, newEnv(t, dir, cfg, nil)), Warn, "not release or prod")

	head := testutil.Git(t, dir, "rev-parse", "HEAD")
	testutil.Git(t, dir, "checkout", "-q", "--detach", head[:len(head)-1])
	expect(t, run(t, GitBranch, newEnv(t, dir, "", nil)), Skip, "Detached HEAD")
}

func TestGitBranchWithoutConventionalBranch(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"a.txt": "a"})
	testutil.GitRepo(t, dir)
	testutil.Git(t, dir, "branch", "-m", "trunk")
	expect(t, run(t, GitBranch, newEnv(t, dir, "", nil)), Skip, "No release branch configured")
}
