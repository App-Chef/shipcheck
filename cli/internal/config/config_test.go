package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var knownIDs = []string{"git.repo", "git.clean", "git.branch", "tests", "build", "environment", "readme", "license", "version", "ci"}

func TestParseFullExample(t *testing.T) {
	cfg, err := Parse([]byte(`
version: 1
checks:
  tests: true
  build: true
  environment: true
  readme: true
  license: false
  ci: false
  git_clean: true
severity:
  readme: error
  git.clean: warning
git:
  protected_branches:
    - main
    - master
environment:
  required:
    - DATABASE_URL
    - API_KEY
  files: [.env, .env.local]
commands:
  tests: go test ./...
  build:
    - npm run build
    - "go build -o 'dist/my app' ./cmd/app"
  timeout: 90s
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.ValidateCheckKeys(knownIDs); err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled("license") || cfg.Enabled("ci") {
		t.Error("license and ci should be disabled")
	}
	if !cfg.Enabled("git.clean") || !cfg.Enabled("version") {
		t.Error("git.clean and unlisted checks should be enabled")
	}
	if got := cfg.SeverityFor("readme"); got != SeverityError {
		t.Errorf("readme severity = %q", got)
	}
	if got := cfg.SeverityFor("git.clean"); got != SeverityWarn {
		t.Errorf("git.clean severity = %q (dotted keys must work)", got)
	}
	if got := cfg.SeverityFor("tests"); got != SeverityDefault {
		t.Errorf("tests severity = %q", got)
	}
	if !reflect.DeepEqual(cfg.Git.ProtectedBranches, []string{"main", "master"}) {
		t.Errorf("branches = %v", cfg.Git.ProtectedBranches)
	}
	if !reflect.DeepEqual(cfg.Environment.Required, []string{"DATABASE_URL", "API_KEY"}) {
		t.Errorf("required = %v", cfg.Environment.Required)
	}
	if !reflect.DeepEqual(cfg.EnvFiles(), []string{".env", ".env.local"}) {
		t.Errorf("env files = %v", cfg.EnvFiles())
	}
	if !reflect.DeepEqual([]string(cfg.Commands.Tests), []string{"go test ./..."}) {
		t.Errorf("tests = %v", cfg.Commands.Tests)
	}
	if len(cfg.Commands.Build) != 2 {
		t.Errorf("build = %v", cfg.Commands.Build)
	}
	if cfg.Timeout() != 90*time.Second {
		t.Errorf("timeout = %v", cfg.Timeout())
	}
}

func TestDefaults(t *testing.T) {
	cfg, err := Parse([]byte("version: 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range knownIDs {
		if !cfg.Enabled(id) {
			t.Errorf("%s should be enabled by default", id)
		}
	}
	if cfg.Timeout() != DefaultTimeout {
		t.Errorf("timeout = %v", cfg.Timeout())
	}
	if !reflect.DeepEqual(cfg.EnvFiles(), []string{".env"}) {
		t.Errorf("env files = %v", cfg.EnvFiles())
	}
}

func TestParseEmptyFileUsesDefaults(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Errorf("version = %d", cfg.Version)
	}
}

func TestParseErrors(t *testing.T) {
	tests := map[string]struct {
		yaml string
		want string
	}{
		"unknown top-level key":  {"version: 1\nchekcs: {}\n", "chekcs"},
		"unknown nested key":     {"version: 1\ngit:\n  branches: [main]\n", "branches"},
		"unsupported version":    {"version: 2\n", "unsupported config version 2"},
		"bad severity":           {"version: 1\nseverity:\n  ci: loud\n", `invalid severity "loud"`},
		"bad env name":           {"version: 1\nenvironment:\n  required: [\"MY-VAR\"]\n", `"MY-VAR" is not a valid variable name`},
		"env file escapes":       {"version: 1\nenvironment:\n  files: [../secrets/.env]\n", "must stay inside the project"},
		"env file absolute":      {"version: 1\nenvironment:\n  files: [/etc/passwd]\n", "must be relative"},
		"bad duration":           {"version: 1\ncommands:\n  timeout: forever\n", `invalid duration "forever"`},
		"negative duration":      {"version: 1\ncommands:\n  timeout: -5s\n", "must be positive"},
		"shell pipeline":         {"version: 1\ncommands:\n  tests: npm test | tee out.log\n", "shell syntax"},
		"command chaining":       {"version: 1\ncommands:\n  build: \"npm ci && npm run build\"\n", "shell syntax"},
		"command substitution":   {"version: 1\ncommands:\n  build: echo $(whoami)\n", "shell syntax"},
		"empty command":          {"version: 1\ncommands:\n  tests: \"  \"\n", "command cannot be empty"},
		"command map":            {"version: 1\ncommands:\n  tests: {run: x}\n", "expected a command string"},
		"empty protected branch": {"version: 1\ngit:\n  protected_branches: [\"\"]\n", "cannot be empty"},
		"invalid yaml":           {"version: 1\nchecks: [\n", "did not find expected"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte(tc.yaml))
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err, tc.want)
			}
		})
	}
}

func TestValidateCheckKeys(t *testing.T) {
	cfg, err := Parse([]byte("version: 1\nchecks:\n  test: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	err = cfg.ValidateCheckKeys(knownIDs)
	if err == nil || !strings.Contains(err.Error(), `unknown check "test"`) || !strings.Contains(err.Error(), "tests") {
		t.Fatalf("got %v", err)
	}

	cfg, _ = Parse([]byte("version: 1\nseverity:\n  lint: warn\n"))
	if err := cfg.ValidateCheckKeys(knownIDs); err == nil || !strings.Contains(err.Error(), "severity") {
		t.Fatalf("got %v", err)
	}
}

func TestSeverityAliases(t *testing.T) {
	cfg, err := Parse([]byte("version: 1\nseverity:\n  ci: WARNING\n  readme: fail\n  license: Warn\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SeverityFor("ci") != SeverityWarn || cfg.SeverityFor("license") != SeverityWarn || cfg.SeverityFor("readme") != SeverityError {
		t.Errorf("severities = %v", cfg.Severity)
	}
}

func TestFindAndLoad(t *testing.T) {
	dir := t.TempDir()
	if Find(dir) != "" {
		t.Fatal("found a config in an empty dir")
	}
	path := filepath.Join(dir, ".shipcheck.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nchecks:\n  ci: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Find(dir) != path {
		t.Fatalf("Find = %q", Find(dir))
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path != path || cfg.Enabled("ci") {
		t.Errorf("cfg = %+v", cfg)
	}

	// .yml wins over .yaml.
	yml := filepath.Join(dir, ".shipcheck.yml")
	if err := os.WriteFile(yml, []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Find(dir) != yml {
		t.Errorf("Find = %q, want .yml", Find(dir))
	}

	if _, err := Load(filepath.Join(dir, "missing.yml")); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("missing file error = %v", err)
	}
	bad := filepath.Join(dir, "bad.yml")
	_ = os.WriteFile(bad, []byte("version: 3\n"), 0o644)
	if _, err := Load(bad); err == nil || !strings.HasPrefix(err.Error(), "bad.yml: ") {
		t.Errorf("error should name the file: %v", err)
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"go test ./...", []string{"go", "test", "./..."}},
		{"  npm   test  ", []string{"npm", "test"}},
		{`pytest -k "slow and not flaky"`, []string{"pytest", "-k", "slow and not flaky"}},
		{`echo 'a|b' "c;d"`, []string{"echo", "a|b", "c;d"}},
		{`go build -o ""`, []string{"go", "build", "-o", ""}},
		{"make\ttest", []string{"make", "test"}},
	}
	for _, tc := range tests {
		got, err := ParseCommand(tc.in)
		if err != nil {
			t.Errorf("ParseCommand(%q): %v", tc.in, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ParseCommand(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	for _, bad := range []string{"", "a && b", "a; b", "a | b", "a > out", "echo $HOME", "echo `id`", "rm *", "a\nb", `echo "unterminated`} {
		if _, err := ParseCommand(bad); err == nil {
			t.Errorf("ParseCommand(%q) should fail", bad)
		}
	}
}

func TestTemplateIsValidConfig(t *testing.T) {
	for _, d := range []TemplateData{
		{},
		{Branch: "develop", EnvTemplate: ".env.example", TestCommands: []string{"go test ./..."}, BuildCommands: []string{"go build ./..."}},
		{Branch: "release/1.0", TestCommands: []string{"go test ./...", "npm test"}},
	} {
		out := Template(d)
		cfg, err := Parse(out)
		if err != nil {
			t.Fatalf("template does not parse: %v\n%s", err, out)
		}
		if err := cfg.ValidateCheckKeys(knownIDs); err != nil {
			t.Fatalf("template has unknown checks: %v", err)
		}
		if d.Branch != "" && !strings.Contains(string(out), d.Branch) {
			t.Errorf("template should include branch %q", d.Branch)
		}
		for _, c := range d.TestCommands {
			if !strings.Contains(string(out), c) {
				t.Errorf("template should mention detected command %q", c)
			}
		}
		if cfg.SeverityFor("ci") != SeverityWarn {
			t.Error("template should keep CI as a warning")
		}
	}
}
