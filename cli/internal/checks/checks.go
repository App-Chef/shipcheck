// Package checks defines Shipcheck's checks.
//
// A check is a small value with an ID, a human name, a description and a
// Run function. To add a check, write a Check in its own file and append
// it to All. The runner, output and configuration pick it up
// automatically: its ID becomes its config key ("git.clean" is configured
// as "git_clean").
package checks

import (
	"context"
	"sync"

	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/detector"
	"github.com/App-Chef/shipcheck/cli/internal/project"
)

// Status is the outcome of a check.
type Status string

const (
	Pass Status = "pass"
	Warn Status = "warn"
	Fail Status = "fail"
	Skip Status = "skip"
)

// Result is what a check reports. Message is the one line shown in the
// terminal ("Working tree clean", "3 uncommitted changes"). Suggestion
// tells the user how to fix a problem. Details are shown with --verbose.
type Result struct {
	Status     Status
	Message    string
	Suggestion string
	Details    []string
}

// Check is a single readiness check.
type Check struct {
	ID          string
	Name        string
	Description string
	Run         func(ctx context.Context, env *Env) Result
}

// All returns every check in the order it runs and is displayed.
func All() []Check {
	return []Check{
		GitRepo,
		GitClean,
		GitBranch,
		Tests,
		Build,
		Environment,
		Readme,
		License,
		Version,
		CI,
	}
}

// IDs returns the IDs of every check.
func IDs() []string {
	all := All()
	ids := make([]string, len(all))
	for i, c := range all {
		ids[i] = c.ID
	}
	return ids
}

// Env is the shared context passed to every check.
type Env struct {
	Project *project.Project
	Config  *config.Config
	Exec    Executor

	detectOnce sync.Once
	detection  detector.Detection

	secretsOnce sync.Once
	secrets     []string
}

// NewEnv builds the environment for a run.
func NewEnv(p *project.Project, cfg *config.Config, exec Executor) *Env {
	if exec == nil {
		exec = SystemExecutor{}
	}
	return &Env{Project: p, Config: cfg, Exec: exec}
}

// Detection returns the (cached) detected ecosystems and commands.
func (e *Env) Detection() detector.Detection {
	e.detectOnce.Do(func() { e.detection = detector.Detect(e.Project) })
	return e.detection
}

func pass(msg string, details ...string) Result {
	return Result{Status: Pass, Message: msg, Details: details}
}

func skip(msg string) Result {
	return Result{Status: Skip, Message: msg}
}
