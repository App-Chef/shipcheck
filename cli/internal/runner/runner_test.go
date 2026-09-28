package runner

import (
	"context"
	"strings"
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/project"
)

func fixed(id string, s checks.Status) checks.Check {
	return checks.Check{
		ID: id, Name: id, Description: "fixed " + string(s),
		Run: func(context.Context, *checks.Env) checks.Result {
			return checks.Result{Status: s, Message: id + " " + string(s), Suggestion: "fix " + id}
		},
	}
}

func env(t *testing.T, yaml string) *checks.Env {
	t.Helper()
	p, err := project.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	if yaml != "" {
		if cfg, err = config.Parse([]byte(yaml)); err != nil {
			t.Fatal(err)
		}
	}
	return checks.NewEnv(p, cfg, nil)
}

type recorder struct{ events []string }

func (r *recorder) Start(c checks.Check) { r.events = append(r.events, "start "+c.ID) }
func (r *recorder) Finish(res Result)    { r.events = append(r.events, "finish "+res.ID) }

func TestRunSummary(t *testing.T) {
	all := []checks.Check{
		fixed("a", checks.Pass), fixed("b", checks.Pass),
		fixed("c", checks.Warn), fixed("d", checks.Skip),
	}
	rec := &recorder{}
	rep, err := Run(context.Background(), all, env(t, ""), rec)
	if err != nil {
		t.Fatal(err)
	}
	want := Summary{Passed: 2, Warnings: 1, Skipped: 1}
	if rep.Summary != want {
		t.Errorf("summary = %+v, want %+v", rep.Summary, want)
	}
	if !rep.Ready() {
		t.Error("warnings and skips must not block shipping")
	}
	if len(rep.Results) != 4 || rep.Results[2].Message != "c warn" || rep.Results[2].Description != "fixed warn" {
		t.Errorf("results = %+v", rep.Results)
	}
	if got := strings.Join(rec.events, ","); got != "start a,finish a,start b,finish b,start c,finish c,start d,finish d" {
		t.Errorf("observer events = %s", got)
	}
}

func TestRunFailureMeansNotReady(t *testing.T) {
	rep, err := Run(context.Background(), []checks.Check{fixed("a", checks.Pass), fixed("b", checks.Fail)}, env(t, ""), nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Ready() || rep.Summary.Failed != 1 {
		t.Errorf("report = %+v", rep.Summary)
	}
}

func TestSeverityOverrides(t *testing.T) {
	all := []checks.Check{fixed("tests", checks.Fail), fixed("ci", checks.Warn), fixed("readme", checks.Pass), fixed("build", checks.Skip)}
	yaml := "version: 1\nseverity:\n  tests: warn\n  ci: error\n  readme: error\n  build: error\n"
	rep, err := Run(context.Background(), all, env(t, yaml), nil)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]checks.Status{}
	for _, r := range rep.Results {
		got[r.ID] = r.Status
	}
	want := map[string]checks.Status{"tests": checks.Warn, "ci": checks.Fail, "readme": checks.Pass, "build": checks.Skip}
	for id, s := range want {
		if got[id] != s {
			t.Errorf("%s = %s, want %s", id, got[id], s)
		}
	}
	if rep.Ready() {
		t.Error("ci upgraded to error should block shipping")
	}
}

func TestDisabledChecksAreOmitted(t *testing.T) {
	all := []checks.Check{fixed("git.clean", checks.Fail), fixed("readme", checks.Pass)}
	rep, err := Run(context.Background(), all, env(t, "version: 1\nchecks:\n  git_clean: false\n"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Results) != 1 || rep.Results[0].ID != "readme" || !rep.Ready() {
		t.Errorf("results = %+v", rep.Results)
	}
}

func TestPanicIsAnError(t *testing.T) {
	boom := checks.Check{ID: "boom", Name: "Boom", Description: "d", Run: func(context.Context, *checks.Env) checks.Result {
		panic("oops")
	}}
	_, err := Run(context.Background(), []checks.Check{boom}, env(t, ""), nil)
	if err == nil || !strings.Contains(err.Error(), "check boom crashed: oops") {
		t.Errorf("err = %v", err)
	}
}

func TestCancelledContextStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Run(ctx, []checks.Check{fixed("a", checks.Pass)}, env(t, ""), nil); err == nil {
		t.Error("expected context error")
	}
}
