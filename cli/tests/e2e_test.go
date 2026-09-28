// Package tests builds the real shipcheck binary and runs it against
// throwaway projects, exactly as a user would.
package tests

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "shipcheck-e2e")
	if err != nil {
		panic(err)
	}
	binary = filepath.Join(dir, "shipcheck")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-ldflags", "-X main.version=v0.0.0-test", "-o", binary, "../cmd/shipcheck")
	if out, err := build.CombinedOutput(); err != nil {
		panic("build shipcheck: " + err.Error() + "\n" + string(out))
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// shipcheck runs the binary in dir and returns exit code and output.
func shipcheck(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return code, stdout.String(), stderr.String()
}

// goProject is a tiny Go module with a real test.
func goProject(t *testing.T, testBody string) string {
	t.Helper()
	dir := testutil.Dir(t, map[string]string{
		"go.mod":                   "module example.com/demo\n\ngo 1.21\n",
		"demo.go":                  "package demo\n\nfunc Add(a, b int) int { return a + b }\n",
		"demo_test.go":             "package demo\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n" + testBody + "\n}\n",
		"README.md":                "# Demo\n",
		"LICENSE":                  "MIT License\n",
		".github/workflows/ci.yml": "on: push\n",
		".gitignore":               ".env\n",
	})
	testutil.GitRepo(t, dir)
	testutil.Git(t, dir, "tag", "v0.1.0")
	return dir
}

func TestReadyProject(t *testing.T) {
	dir := goProject(t, `if Add(2, 2) != 4 { t.Fatal("math") }`)
	code, out, stderr := shipcheck(t, dir)
	if code != 0 {
		t.Fatalf("exit %d\n%s\n%s", code, out, stderr)
	}
	for _, want := range []string{
		"SHIPCHECK",
		"✓ Git repository",
		"✓ Working tree clean",
		"✓ On branch main",
		"✓ Tests passed",
		"✓ Production build passed",
		"✓ README found",
		"✓ License found (MIT)",
		"✓ Version v0.1.0",
		"✓ CI configured (GitHub Actions)",
		"9 passed · 1 skipped",
		"✓ Ready to ship",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Error("NO_COLOR output contains ANSI escapes")
	}

	// `shipcheck && deploy` style chaining.
	if code, _, _ := shipcheck(t, dir, "check"); code != 0 {
		t.Errorf("shipcheck check exit %d", code)
	}
}

func TestFailingTests(t *testing.T) {
	dir := goProject(t, `t.Fatal("deliberately broken")`)
	code, out, _ := shipcheck(t, dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1\n%s", code, out)
	}
	if !strings.Contains(out, "✗ Tests failed (go test ./...)") || !strings.Contains(out, "✗ Not ready to ship") {
		t.Errorf("output:\n%s", out)
	}

	code, out, _ = shipcheck(t, dir, "--verbose")
	if code != 1 || !strings.Contains(out, "deliberately broken") {
		t.Errorf("--verbose should include the failing test output:\n%s", out)
	}
}

func TestJSONIsTheOnlyOutput(t *testing.T) {
	dir := goProject(t, `t.Fatal("broken")`)
	testutil.Write(t, dir, map[string]string{"scratch.txt": "uncommitted"})

	code, out, stderr := shipcheck(t, dir, "--json")
	if code != 1 {
		t.Errorf("exit %d", code)
	}
	if stderr != "" {
		t.Errorf("stderr must be empty in --json mode: %q", stderr)
	}
	var report struct {
		Ready   bool `json:"ready"`
		Summary struct {
			Passed   int `json:"passed"`
			Warnings int `json:"warnings"`
			Failed   int `json:"failed"`
			Skipped  int `json:"skipped"`
		} `json:"summary"`
		Checks []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if report.Ready || report.Summary.Failed != 2 {
		t.Errorf("report: %+v", report)
	}
	got := map[string]string{}
	for _, c := range report.Checks {
		got[c.ID] = c.Status
	}
	if got["git.clean"] != "fail" || got["tests"] != "fail" || got["readme"] != "pass" {
		t.Errorf("statuses: %v", got)
	}
}

func TestEnvironmentNeverPrintsValues(t *testing.T) {
	dir := goProject(t, `if Add(1, 1) != 2 { t.Fatal() }`)
	testutil.Write(t, dir, map[string]string{".env.example": "SC_E2E_TOKEN=\nSC_E2E_MISSING=\n"})
	testutil.Commit(t, dir, "env template")
	testutil.Write(t, dir, map[string]string{".env": "SC_E2E_TOKEN=tok_very_secret_value\n"})

	for _, args := range [][]string{nil, {"--verbose"}, {"--json", "--verbose"}} {
		code, out, stderr := shipcheck(t, dir, args...)
		if code != 1 {
			t.Errorf("%v: exit %d", args, code)
		}
		if strings.Contains(out+stderr, "tok_very_secret_value") {
			t.Fatalf("%v: secret value printed:\n%s", args, out)
		}
		if !strings.Contains(out, "1 environment variable missing") {
			t.Errorf("%v: output:\n%s", args, out)
		}
	}
}

func TestInitThenCheck(t *testing.T) {
	dir := goProject(t, `if Add(1, 2) != 3 { t.Fatal() }`)
	code, out, _ := shipcheck(t, dir, "init")
	if code != 0 || !strings.Contains(out, "Created .shipcheck.yml") {
		t.Fatalf("init: %d %s", code, out)
	}
	testutil.Commit(t, dir, "add shipcheck config")
	if code, out, _ := shipcheck(t, dir); code != 0 {
		t.Errorf("check after init: exit %d\n%s", code, out)
	}
	if code, _, stderr := shipcheck(t, dir, "init"); code != 2 || !strings.Contains(stderr, "already exists") {
		t.Errorf("second init should refuse: %d %q", code, stderr)
	}
}

func TestConfiguredCommandInSubdirectory(t *testing.T) {
	root := t.TempDir()
	testutil.Write(t, root, map[string]string{
		"README.md":           "# Mono\n",
		"service/go.mod":      "module example.com/svc\n\ngo 1.21\n",
		"service/svc.go":      "package svc\n",
		"service/svc_test.go": "package svc\n\nimport \"testing\"\n\nfunc TestOK(t *testing.T) {}\n",
		".shipcheck.yml":      "version: 1\ncommands:\n  tests: go -C service test ./...\n  build: go -C service build ./...\n",
	})
	testutil.GitRepo(t, root)

	code, out, _ := shipcheck(t, root, "--verbose")
	if code != 0 || !strings.Contains(out, "$ go -C service test ./...") {
		t.Errorf("exit %d\n%s", code, out)
	}
}

func TestErrorsExitTwo(t *testing.T) {
	dir := t.TempDir()
	testutil.Write(t, dir, map[string]string{".shipcheck.yml": "version: 1\ncommands:\n  tests: npm test && rm -rf /\n"})
	code, _, stderr := shipcheck(t, dir)
	if code != 2 || !strings.Contains(stderr, "shell syntax") {
		t.Errorf("exit %d, stderr %q", code, stderr)
	}
	if code, _, _ := shipcheck(t, dir, "--not-a-flag"); code != 2 {
		t.Errorf("bad flag exit %d", code)
	}
}

func TestVersionAndHelp(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{"version"}, {"--version"}} {
		code, out, _ := shipcheck(t, dir, args...)
		if code != 0 || strings.TrimSpace(out) != "shipcheck v0.0.0-test" {
			t.Errorf("%v: %d %q", args, code, out)
		}
	}
	code, out, _ := shipcheck(t, dir, "help")
	if code != 0 || !strings.Contains(out, "Know you're ready before you ship.") {
		t.Errorf("help: %d\n%s", code, out)
	}
}
