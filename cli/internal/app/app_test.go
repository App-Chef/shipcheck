package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

// stubExec pretends every command succeeds unless listed in fail.
type stubExec struct{ fail map[string]bool }

func (s stubExec) Run(_ context.Context, _ string, args []string, _ []string) checks.ExecResult {
	if s.fail[strings.Join(args, " ")] {
		return checks.ExecResult{ExitCode: 1, Output: []byte("boom")}
	}
	return checks.ExecResult{}
}

func useExec(t *testing.T, e checks.Executor) {
	t.Helper()
	old := executor
	executor = e
	t.Cleanup(func() { executor = old })
}

func execute(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = Execute(args, &out, &errOut, BuildInfo{Version: "v9.9.9"})
	return code, out.String(), errOut.String()
}

// readyProject is a committed repo that passes every check.
func readyProject(t *testing.T) string {
	t.Helper()
	dir := testutil.Dir(t, map[string]string{
		"README.md":                "# App\n",
		"LICENSE":                  "MIT License\n",
		"VERSION":                  "1.0.0\n",
		".github/workflows/ci.yml": "on: push\n",
		"go.mod":                   "module example.com/app\n",
	})
	testutil.GitRepo(t, dir)
	return dir
}

func TestDefaultCommandIsCheck(t *testing.T) {
	useExec(t, stubExec{})
	dir := readyProject(t)

	code1, out1, _ := execute("--dir", dir)
	code2, out2, _ := execute("check", "--dir", dir)
	if code1 != ExitReady || code2 != ExitReady {
		t.Fatalf("exit codes %d %d\n%s", code1, code2, out1)
	}
	if out1 != out2 {
		t.Errorf("shipcheck and shipcheck check differ:\n%s\n---\n%s", out1, out2)
	}
	for _, want := range []string{"SHIPCHECK", "✓ Tests passed", "✓ Production build passed", "✓ Ready to ship"} {
		if !strings.Contains(out1, want) {
			t.Errorf("output missing %q:\n%s", want, out1)
		}
	}
}

func TestExitCodeOneOnFailure(t *testing.T) {
	useExec(t, stubExec{fail: map[string]bool{"go test ./...": true}})
	dir := readyProject(t)
	code, out, _ := execute("--dir", dir)
	if code != ExitNotReady {
		t.Fatalf("exit = %d, want 1\n%s", code, out)
	}
	if !strings.Contains(out, "✗ Tests failed (go test ./...)") || !strings.Contains(out, "✗ Not ready to ship") {
		t.Errorf("output:\n%s", out)
	}
	if strings.Contains(out, "boom") {
		t.Error("command output must only appear with --verbose")
	}

	_, verbose, _ := execute("--dir", dir, "--verbose")
	if !strings.Contains(verbose, "boom") {
		t.Errorf("--verbose should show command output:\n%s", verbose)
	}
}

func TestWarningsStillReady(t *testing.T) {
	useExec(t, stubExec{})
	dir := testutil.Dir(t, map[string]string{"README.md": "# x\n"})
	testutil.GitRepo(t, dir)
	code, out, _ := execute("--dir", dir)
	if code != ExitReady || !strings.Contains(out, "⚠ No CI configuration") {
		t.Errorf("exit %d:\n%s", code, out)
	}
}

func TestJSONOutput(t *testing.T) {
	useExec(t, stubExec{fail: map[string]bool{"go test ./...": true}})
	dir := readyProject(t)
	code, out, stderr := execute("--dir", dir, "--json")
	if code != ExitNotReady {
		t.Errorf("exit = %d", code)
	}
	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}
	var report struct {
		Ready   bool   `json:"ready"`
		Version string `json:"version"`
		Summary struct {
			Passed, Warnings, Failed, Skipped int
		} `json:"summary"`
		Checks []struct {
			ID, Status string
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("stdout is not pure JSON: %v\n%s", err, out)
	}
	if report.Ready || report.Version != "v9.9.9" || report.Summary.Failed != 1 {
		t.Errorf("report = %+v", report)
	}
	statuses := map[string]string{}
	for _, c := range report.Checks {
		statuses[c.ID] = c.Status
	}
	if statuses["tests"] != "fail" || statuses["git.clean"] != "pass" || len(statuses) != len(checks.All()) {
		t.Errorf("statuses = %v", statuses)
	}
}

func TestConfigDisablesAndOverrides(t *testing.T) {
	useExec(t, stubExec{fail: map[string]bool{"go test ./...": true}})
	dir := readyProject(t)
	cfg := "version: 1\nchecks:\n  build: false\nseverity:\n  tests: warn\n"
	testutil.Write(t, dir, map[string]string{".shipcheck.yml": cfg})
	testutil.Commit(t, dir, "config")

	code, out, _ := execute("--dir", dir)
	if code != ExitReady {
		t.Errorf("tests downgraded to warn should be ready, exit %d\n%s", code, out)
	}
	if strings.Contains(out, "Production build") {
		t.Error("disabled build check should not appear")
	}
	if !strings.Contains(out, "⚠ Tests failed") {
		t.Errorf("output:\n%s", out)
	}
}

func TestExplicitConfigFlag(t *testing.T) {
	useExec(t, stubExec{})
	dir := readyProject(t)
	cfgPath := filepath.Join(t.TempDir(), "ci.yml")
	testutil.Write(t, filepath.Dir(cfgPath), map[string]string{"ci.yml": "version: 1\nchecks:\n  tests: false\n"})

	_, out, _ := execute("--dir", dir, "--config", cfgPath, "--verbose")
	if strings.Contains(out, "Tests passed") || !strings.Contains(out, "config   "+cfgPath) {
		t.Errorf("output:\n%s", out)
	}
}

func TestShipcheckErrorsExitTwo(t *testing.T) {
	useExec(t, stubExec{})
	dir := readyProject(t)
	testutil.Write(t, dir, map[string]string{".shipcheck.yml": "version: 1\nchecks:\n  test: false\n"})

	cases := [][]string{
		{"--dir", dir},                               // unknown check key
		{"--dir", filepath.Join(dir, "nope")},        // missing directory
		{"--config", filepath.Join(dir, "none.yml")}, // missing config
		{"--bogus-flag"},
		{"unknown-command"},
	}
	for _, args := range cases {
		code, out, stderr := execute(args...)
		if code != ExitError {
			t.Errorf("%v: exit = %d, want 2\n%s", args, code, out)
		}
		if !strings.HasPrefix(stderr, "shipcheck: ") {
			t.Errorf("%v: stderr = %q", args, stderr)
		}
	}

	// JSON consumers get JSON errors on stdout, even for flag errors.
	for _, args := range [][]string{{"--dir", dir, "--json"}, {"--json", "--bogus-flag"}} {
		code, out, stderr := execute(args...)
		var got map[string]string
		if code != ExitError || json.Unmarshal([]byte(out), &got) != nil || got["error"] == "" || stderr != "" {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", args, code, out, stderr)
		}
	}
}

func TestInit(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"go.mod": "module x\n", ".env.example": "API_KEY=\n"})

	code, out, _ := execute("init", "--dir", dir)
	if code != ExitReady || !strings.Contains(out, "Created .shipcheck.yml") {
		t.Fatalf("exit %d: %s", code, out)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".shipcheck.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "go test ./...") || !strings.Contains(string(data), ".env.example") {
		t.Errorf("generated config should mention detected commands and env template:\n%s", data)
	}
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatalf("generated config is invalid: %v", err)
	}
	if err := cfg.ValidateCheckKeys(checks.IDs()); err != nil {
		t.Fatal(err)
	}

	// Never overwrite without --force.
	code, _, stderr := execute("init", "--dir", dir)
	if code != ExitError || !strings.Contains(stderr, "already exists") {
		t.Errorf("second init: exit %d, %q", code, stderr)
	}
	if code, _, _ := execute("init", "--dir", dir, "--force"); code != ExitReady {
		t.Errorf("init --force exit %d", code)
	}
}

func TestVersion(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}} {
		code, out, _ := execute(args...)
		if code != 0 || out != "shipcheck v9.9.9\n" {
			t.Errorf("%v: %d %q", args, code, out)
		}
	}
	_, out, _ := execute("version", "--json")
	var v map[string]string
	if err := json.Unmarshal([]byte(out), &v); err != nil || v["version"] != "v9.9.9" {
		t.Errorf("version --json = %q", out)
	}
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"--help"}, {"help", "init"}} {
		code, out, _ := execute(args...)
		if code != 0 || !strings.Contains(out, "shipcheck") {
			t.Errorf("%v: exit %d\n%s", args, code, out)
		}
	}
	_, out, _ := execute("help")
	for _, want := range []string{"check", "init", "version", "--json", "--verbose", "--config"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestResolveBuildInfo(t *testing.T) {
	b := ResolveBuildInfo(BuildInfo{Version: "v1.2.3", Commit: "0123456789abcdef0123"})
	if b.Version != "v1.2.3" || b.Commit != "0123456789ab" {
		t.Errorf("got %+v", b)
	}
	if ResolveBuildInfo(BuildInfo{}).Version == "" {
		t.Error("version must never be empty")
	}
}
