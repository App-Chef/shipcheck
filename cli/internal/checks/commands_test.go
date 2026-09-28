package checks

import (
	"context"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

func TestTestsSkipWhenNothingDetected(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"README.md": "hi"})
	res := run(t, Tests, newEnv(t, dir, "", nil))
	expect(t, res, Skip, "No test command detected")
	if !strings.Contains(res.Suggestion, "commands.tests") {
		t.Errorf("suggestion = %q", res.Suggestion)
	}
	expect(t, run(t, Build, newEnv(t, dir, "", nil)), Skip, "No build command detected")
}

func TestTestsPassUsingDetectedCommand(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"go.mod": "module x\n"})
	fake := &fakeExec{results: map[string]ExecResult{}}
	res := run(t, Tests, newEnv(t, dir, "", fake))
	expect(t, res, Pass, "Tests passed")
	if !slices.Equal(fake.ran, []string{"go test ./..."}) {
		t.Errorf("ran %q", fake.ran)
	}
	if !slices.Contains(fake.env[0], "CI=true") {
		t.Errorf("tests should run with CI=true, env = %q", fake.env[0])
	}
	if len(res.Details) != 1 || !strings.HasPrefix(res.Details[0], "$ go test ./...") {
		t.Errorf("details = %q", res.Details)
	}
}

func TestBuildFailure(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"package.json": `{"scripts":{"build":"next build"}}`})
	var output strings.Builder
	for i := 1; i <= 30; i++ {
		output.WriteString("line ")
		output.WriteString(strings.Repeat("x", i%3))
		output.WriteString("\n")
	}
	output.WriteString("Error: Module not found\n")
	fake := &fakeExec{results: map[string]ExecResult{
		"npm run build": {ExitCode: 1, Output: []byte(output.String())},
	}}
	res := run(t, Build, newEnv(t, dir, "", fake))
	expect(t, res, Fail, "Production build failed (npm run build)")
	if !strings.Contains(res.Suggestion, "`npm run build`") {
		t.Errorf("suggestion = %q", res.Suggestion)
	}
	// One header, one "hidden" marker and the last 20 lines.
	if len(res.Details) != 22 {
		t.Fatalf("got %d detail lines: %q", len(res.Details), res.Details)
	}
	if !strings.Contains(res.Details[1], "11 earlier lines hidden") || res.Details[21] != "Error: Module not found" {
		t.Errorf("details = %q", res.Details)
	}
	if slices.Contains(fake.env[0], "CI=true") {
		t.Error("builds should not force CI=true")
	}
}

func TestCommandNotInstalled(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"Cargo.toml": "[package]\n"})
	fake := &fakeExec{results: map[string]ExecResult{"cargo test": {NotFound: true, ExitCode: -1}}}
	res := run(t, Tests, newEnv(t, dir, "", fake))
	expect(t, res, Fail, "cargo is not installed")
}

func TestCommandTimeout(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"go.mod": "module x\n"})
	fake := &fakeExec{results: map[string]ExecResult{"go test ./...": {TimedOut: true, ExitCode: -1}}}
	res := run(t, Tests, newEnv(t, dir, "version: 1\ncommands:\n  timeout: 2m\n", fake))
	expect(t, res, Fail, "go test ./... timed out after 2m0s")
}

func TestConfiguredCommandsOverrideDetection(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"go.mod": "module x\n"})
	fake := &fakeExec{results: map[string]ExecResult{
		"make test":           {},
		"go vet ./...":        {},
		"npm run test:e2e -w": {ExitCode: 2},
	}}
	cfg := "version: 1\ncommands:\n  tests:\n    - make test\n    - go vet ./...\n"
	expect(t, run(t, Tests, newEnv(t, dir, cfg, fake)), Pass, "Tests passed")
	if !slices.Equal(fake.ran, []string{"make test", "go vet ./..."}) {
		t.Errorf("ran %q", fake.ran)
	}

	// Stops at the first failing command.
	fake.ran = nil
	cfg = "version: 1\ncommands:\n  tests: [\"npm run test:e2e -w\", make test]\n"
	expect(t, run(t, Tests, newEnv(t, dir, cfg, fake)), Fail, "Tests failed (npm run test:e2e -w)")
	if len(fake.ran) != 1 {
		t.Errorf("ran %q", fake.ran)
	}
}

func TestCommandOutputIsRedacted(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{
		"go.mod":       "module x\n",
		".env.example": "API_KEY=\n",
		".env":         "API_KEY=sk_live_supersecret\n",
	})
	fake := &fakeExec{results: map[string]ExecResult{
		"go test ./...": {ExitCode: 1, Output: []byte("connecting with sk_live_supersecret\nFAIL")},
	}}
	res := run(t, Tests, newEnv(t, dir, "", fake))
	joined := strings.Join(res.Details, "\n")
	if strings.Contains(joined, "supersecret") || !strings.Contains(joined, "connecting with ****") {
		t.Errorf("secret leaked or not redacted: %q", joined)
	}
}

// TestSystemExecutor exercises real processes using the Go toolchain,
// which is always present when these tests run.
func TestSystemExecutor(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH")
	}
	ex := SystemExecutor{}
	dir := t.TempDir()

	res := ex.Run(context.Background(), dir, []string{goBin, "version"}, nil)
	if res.ExitCode != 0 || res.Err != nil || !strings.Contains(string(res.Output), "go version") {
		t.Errorf("go version: %+v", res)
	}

	res = ex.Run(context.Background(), dir, []string{goBin, "no-such-subcommand"}, nil)
	if res.ExitCode == 0 {
		t.Errorf("expected non-zero exit, got %+v", res)
	}

	res = ex.Run(context.Background(), dir, []string{"shipcheck-definitely-not-installed"}, nil)
	if !res.NotFound {
		t.Errorf("expected NotFound, got %+v", res)
	}

	res = ex.Run(context.Background(), dir, nil, nil)
	if res.Err == nil {
		t.Error("expected an error for an empty command")
	}
}

func TestSystemExecutorTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX sleep binary")
	}
	sleep, err := exec.LookPath("sleep")
	if err != nil {
		t.Skip("sleep not available")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	res := SystemExecutor{}.Run(ctx, t.TempDir(), []string{sleep, "30"}, nil)
	if !res.TimedOut {
		t.Errorf("expected timeout, got %+v", res)
	}
	if time.Since(start) > 10*time.Second {
		t.Error("timeout did not stop the process promptly")
	}
}

func TestTailBufferKeepsTheEnd(t *testing.T) {
	b := &tailBuffer{max: 10}
	_, _ = b.Write([]byte("0123456789"))
	_, _ = b.Write([]byte("abc"))
	if got := string(b.Bytes()); got != "3456789abc" {
		t.Errorf("got %q", got)
	}
}
