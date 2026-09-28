package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/detector"
)

// Tests runs the project's test suite.
var Tests = Check{
	ID:          "tests",
	Name:        "Tests",
	Description: "Detects and runs the project's test command (go test, npm test, pytest, cargo test, …).",
	Run: func(ctx context.Context, env *Env) Result {
		return runCommands(ctx, env, commandKind{
			label:      "Tests",
			noun:       "test command",
			configured: env.Config.Commands.Tests,
			detected:   env.Detection().Tests,
			configKey:  "commands.tests",
			// Test runners such as Jest and Vitest skip watch mode in CI.
			extraEnv: []string{"CI=true"},
		})
	},
}

// Build runs the project's production build.
var Build = Check{
	ID:          "build",
	Name:        "Production build",
	Description: "Detects and runs the project's production build command (go build, npm run build, cargo build --release, …).",
	Run: func(ctx context.Context, env *Env) Result {
		return runCommands(ctx, env, commandKind{
			label:      "Production build",
			noun:       "build command",
			configured: env.Config.Commands.Build,
			detected:   env.Detection().Builds,
			configKey:  "commands.build",
		})
	},
}

type commandKind struct {
	label      string
	noun       string
	configured config.CommandList
	detected   []detector.Command
	configKey  string
	extraEnv   []string
}

// tailLines is how many lines of output --verbose shows for a failure.
const tailLines = 20

func runCommands(ctx context.Context, env *Env, k commandKind) Result {
	commands, err := resolveCommands(k)
	if err != nil {
		return Result{Status: Fail, Message: "Invalid " + k.noun, Details: []string{err.Error()}}
	}
	if len(commands) == 0 {
		return Result{
			Status:     Skip,
			Message:    "No " + k.noun + " detected",
			Suggestion: fmt.Sprintf("Set %s in .shipcheck.yml if the project has one.", k.configKey),
		}
	}

	timeout := env.Config.Timeout()
	var details []string
	for _, args := range commands {
		display := strings.Join(args, " ")
		cmdCtx, cancel := context.WithTimeout(ctx, timeout)
		start := time.Now()
		res := env.Exec.Run(cmdCtx, env.Project.Root, args, k.extraEnv)
		elapsed := time.Since(start)
		cancel()

		switch {
		case res.NotFound:
			return Result{
				Status:     Fail,
				Message:    fmt.Sprintf("%s not run: %s is not installed", k.label, args[0]),
				Suggestion: fmt.Sprintf("Install %s or set %s in .shipcheck.yml.", args[0], k.configKey),
				Details:    details,
			}
		case res.TimedOut:
			return Result{
				Status:     Fail,
				Message:    fmt.Sprintf("%s timed out after %s", display, timeout),
				Suggestion: "Raise commands.timeout in .shipcheck.yml if this is expected.",
				Details:    append(details, env.tail(res.Output)...),
			}
		case res.Err != nil:
			return Result{
				Status:  Fail,
				Message: fmt.Sprintf("Could not run %s", display),
				Details: append(details, res.Err.Error()),
			}
		case res.ExitCode != 0:
			return Result{
				Status:     Fail,
				Message:    fmt.Sprintf("%s failed (%s)", k.label, display),
				Suggestion: fmt.Sprintf("Run `%s` to see the full output.", display),
				Details:    append(append(details, fmt.Sprintf("$ %s  (exit %d, %s)", display, res.ExitCode, roundDuration(elapsed))), env.tail(res.Output)...),
			}
		}
		details = append(details, fmt.Sprintf("$ %s  (%s)", display, roundDuration(elapsed)))
	}
	return pass(k.label+" passed", details...)
}

func resolveCommands(k commandKind) ([][]string, error) {
	if len(k.configured) > 0 {
		out := make([][]string, 0, len(k.configured))
		for _, c := range k.configured {
			args, err := config.ParseCommand(c)
			if err != nil {
				return nil, err
			}
			out = append(out, args)
		}
		return out, nil
	}
	out := make([][]string, 0, len(k.detected))
	for _, c := range k.detected {
		out = append(out, c.Args)
	}
	return out, nil
}

// tail returns the last lines of command output with secrets redacted.
func (e *Env) tail(output []byte) []string {
	text := strings.TrimRight(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n")
	if text == "" {
		return nil
	}
	lines := strings.Split(e.Redact(text), "\n")
	if len(lines) > tailLines {
		lines = append([]string{fmt.Sprintf("… %d earlier lines hidden", len(lines)-tailLines)}, lines[len(lines)-tailLines:]...)
	}
	return lines
}

func roundDuration(d time.Duration) time.Duration {
	if d < time.Second {
		return d.Round(time.Millisecond)
	}
	return d.Round(100 * time.Millisecond)
}
