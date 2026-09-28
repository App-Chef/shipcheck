# Contributing to Shipcheck

Thanks for helping. Shipcheck is small on purpose: the best contributions make existing checks more accurate, clearer or faster. For new checks or larger changes, please open an issue first so we can agree on scope before you write code.

## Repository layout

```text
cli/                 Go CLI
  cmd/shipcheck/     entry point
  internal/app/      commands, flags, exit codes
  internal/checks/   one file per check
  internal/config/   .shipcheck.yml parsing and validation
  internal/detector/ test, build and version detection
  internal/output/   text and JSON renderers
  internal/project/  filesystem and Git helpers
  internal/runner/   runs checks, applies severity
  tests/             end-to-end tests against the real binary
web/                 website (Next.js, static)
docs/                Markdown documentation
```

## CLI

Requires Go 1.22+ and Git.

```bash
cd cli
go test ./...
go vet ./...
gofmt -l .                              # should print nothing
go run ./cmd/shipcheck --dir .. --verbose
```

Tests build real temporary projects and Git repositories. They need no network and no fixtures on disk. Every behaviour change needs a test.

### Adding a check

A check is a value with an ID, a name, a description and a `Run` function. The runner handles enabling, timing, severity and output.

```go
// cli/internal/checks/changelog.go
package checks

import "context"

var Changelog = Check{
	ID:          "changelog",
	Name:        "Changelog",
	Description: "Checks for a CHANGELOG.md.",
	Run: func(_ context.Context, env *Env) Result {
		if name := env.Project.FindFile("CHANGELOG.md"); name != "" {
			return pass("Changelog found", name)
		}
		return Result{
			Status:     Warn,
			Message:    "No changelog",
			Suggestion: "Add a CHANGELOG.md so users know what changed.",
		}
	},
}
```

1. Add it to `All()` in `cli/internal/checks/checks.go`, in the position it should run. Its config key (`checks.changelog`) works automatically.
2. Add tests next to it. Use `testutil.Dir` for throwaway projects and `testutil.GitRepo` for repositories.
3. Document it in `docs/checks.md` and `web/lib/checks.ts`.

Guidelines:

- Offline and fast. No network access, ever.
- Read-only. Never modify the project.
- Never print secret values.
- Messages describe the result ("No changelog"); suggestions say what to do.
- Fail only for real release blockers. Warn for everything else.
- External commands go through `env.Exec` (so tests can fake them) and never through a shell.

## Website

Requires Node.js 20.9+.

```bash
cd web
npm install
npm run dev
npm run lint
npm run build
```

The site is static: no database, API routes, analytics or tracking. Terminal examples in `web/lib/terminal.ts` are copied from real CLI output; update them when CLI messages change.

## Pull requests

- Keep each pull request focused on one change.
- Make sure CI passes: Go tests on Linux, macOS and Windows, plus lint and build for the website.
- Update documentation (`README.md`, `docs/`, `web/`) when behaviour changes.
- Don't add features outside Shipcheck's scope (pre-shipping validation). Accounts, dashboards, deployment and monitoring are out of scope.

## Releasing (maintainers)

The Go module lives in `cli/`, so release tags carry the directory prefix:

```bash
git tag cli/v0.1.0
git push origin cli/v0.1.0
```

`go install github.com/App-Chef/shipcheck/cli/cmd/shipcheck@latest` then resolves to that version.

## Code of conduct

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold it.
