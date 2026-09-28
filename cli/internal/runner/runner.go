// Package runner executes checks, applies configured severity and
// summarises the results.
package runner

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/config"
)

// Result is a finished check.
type Result struct {
	ID          string
	Name        string
	Description string
	Status      checks.Status
	Message     string
	Suggestion  string
	Details     []string
	Duration    time.Duration
}

// Summary counts results by status.
type Summary struct {
	Passed   int
	Warnings int
	Failed   int
	Skipped  int
}

// Report is the outcome of a full run.
type Report struct {
	Results  []Result
	Summary  Summary
	Duration time.Duration
}

// Ready reports whether nothing failed. Warnings do not block shipping.
func (r *Report) Ready() bool {
	return r.Summary.Failed == 0
}

// Observer is notified as checks start and finish, so output can stream.
type Observer interface {
	Start(c checks.Check)
	Finish(r Result)
}

// Run executes every enabled check in order. Disabled checks are left out
// of the report entirely. A panicking check is a Shipcheck bug and is
// returned as an error rather than reported as a project failure.
func Run(ctx context.Context, all []checks.Check, env *checks.Env, obs Observer) (*Report, error) {
	report := &Report{}
	start := time.Now()
	for _, c := range all {
		if !env.Config.Enabled(c.ID) {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if obs != nil {
			obs.Start(c)
		}
		res, err := runOne(ctx, c, env)
		if err != nil {
			return nil, err
		}
		res.Status = applySeverity(res.Status, env.Config.SeverityFor(c.ID))
		report.Results = append(report.Results, res)
		switch res.Status {
		case checks.Pass:
			report.Summary.Passed++
		case checks.Warn:
			report.Summary.Warnings++
		case checks.Fail:
			report.Summary.Failed++
		case checks.Skip:
			report.Summary.Skipped++
		}
		if obs != nil {
			obs.Finish(res)
		}
	}
	report.Duration = time.Since(start)
	return report, nil
}

func runOne(ctx context.Context, c checks.Check, env *checks.Env) (res Result, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("check %s crashed: %v\n%s", c.ID, p, debug.Stack())
		}
	}()
	start := time.Now()
	out := c.Run(ctx, env)
	return Result{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Status:      out.Status,
		Message:     out.Message,
		Suggestion:  out.Suggestion,
		Details:     out.Details,
		Duration:    time.Since(start),
	}, nil
}

// applySeverity maps a status through the configured severity:
// "warn" downgrades failures, "error" upgrades warnings.
func applySeverity(s checks.Status, sev config.Severity) checks.Status {
	switch {
	case sev == config.SeverityWarn && s == checks.Fail:
		return checks.Warn
	case sev == config.SeverityError && s == checks.Warn:
		return checks.Fail
	}
	return s
}
