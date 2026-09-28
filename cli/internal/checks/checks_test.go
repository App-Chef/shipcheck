package checks

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/project"
)

// fakeExec records commands and returns canned results keyed by command.
type fakeExec struct {
	mu      sync.Mutex
	results map[string]ExecResult
	ran     []string
	env     [][]string
}

func (f *fakeExec) Run(_ context.Context, _ string, args []string, env []string) ExecResult {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := strings.Join(args, " ")
	f.ran = append(f.ran, cmd)
	f.env = append(f.env, env)
	return f.results[cmd]
}

func newEnv(t *testing.T, dir string, cfgYAML string, exec Executor) *Env {
	t.Helper()
	p, err := project.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	if cfgYAML != "" {
		if cfg, err = config.Parse([]byte(cfgYAML)); err != nil {
			t.Fatal(err)
		}
	}
	if exec == nil {
		exec = &fakeExec{}
	}
	return NewEnv(p, cfg, exec)
}

func run(t *testing.T, c Check, env *Env) Result {
	t.Helper()
	return c.Run(context.Background(), env)
}

func expect(t *testing.T, got Result, status Status, msgContains string) {
	t.Helper()
	if got.Status != status {
		t.Errorf("status = %s, want %s (message %q)", got.Status, status, got.Message)
	}
	if !strings.Contains(got.Message, msgContains) {
		t.Errorf("message %q does not contain %q", got.Message, msgContains)
	}
}

func TestRegistry(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range All() {
		if c.ID == "" || c.Name == "" || c.Description == "" || c.Run == nil {
			t.Errorf("check %+v is incomplete", c)
		}
		if seen[c.ID] {
			t.Errorf("duplicate check ID %s", c.ID)
		}
		seen[c.ID] = true
	}
	if len(IDs()) != len(All()) {
		t.Error("IDs() and All() disagree")
	}
}
