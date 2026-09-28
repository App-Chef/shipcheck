package app

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/output"
	"github.com/App-Chef/shipcheck/cli/internal/project"
	"github.com/App-Chef/shipcheck/cli/internal/runner"
)

// executor is swapped in tests to avoid running real test suites.
var executor checks.Executor = checks.SystemExecutor{}

func runCheck(ctx context.Context, opts *options, stdout io.Writer, build BuildInfo) error {
	proj, cfg, err := load(opts)
	if err != nil {
		return err
	}
	env := checks.NewEnv(proj, cfg, executor)
	all := checks.All()

	if opts.json {
		report, err := runner.Run(ctx, all, env, nil)
		if err != nil {
			return interrupted(err)
		}
		if err := output.WriteJSON(stdout, report, build.Version, opts.verbose); err != nil {
			return err
		}
		if !report.Ready() {
			return errNotReady
		}
		return nil
	}

	text := &output.Text{W: stdout, Style: styleFor(stdout, opts.noColor), Verbose: opts.verbose}
	text.Header(proj.Root, cfg.Path)
	report, err := runner.Run(ctx, all, env, text)
	if err != nil {
		return interrupted(err)
	}
	text.Summary(report)
	if !report.Ready() {
		return errNotReady
	}
	return nil
}

// load opens the project and its configuration.
func load(opts *options) (*project.Project, *config.Config, error) {
	proj, err := project.Open(opts.dir)
	if err != nil {
		return nil, nil, err
	}
	path := opts.configPath
	if path == "" {
		path = config.Find(proj.Root)
	} else if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}
	cfg := config.Default()
	if path != "" {
		if cfg, err = config.Load(path); err != nil {
			return nil, nil, err
		}
	}
	if err := cfg.ValidateCheckKeys(checks.IDs()); err != nil {
		return nil, nil, err
	}
	return proj, cfg, nil
}

func styleFor(w io.Writer, noColor bool) output.Style {
	f, ok := w.(*os.File)
	if !ok {
		return output.Style{}
	}
	output.PrepareConsole()
	return output.DetectStyle(f, noColor)
}

func interrupted(err error) error {
	if errors.Is(err, context.Canceled) {
		return errors.New("interrupted")
	}
	return err
}
