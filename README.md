# Shipcheck

Know you're ready before you ship.

A minimal open-source CLI that checks your project
for common release blockers.

```bash
shipcheck
```

```text
SHIPCHECK

✓ Git repository
✓ Working tree clean
✓ On branch main
✓ Tests passed
✓ Production build passed
✓ Environment checked
✓ README found
✓ License found (MIT)
✓ Version 1.4.0
⚠ No CI configuration
  → Add a CI workflow so every push is tested (shipcheck runs in CI too).

────────────────────────────

9 passed · 1 warning

✓ Ready to ship
```

**Website:** _not deployed yet. Replace this line with the site URL once `web/` is live._
Until then, run the site locally (see [Development](#development)) or read [`docs/`](docs/).

---

## What it is

Shipcheck answers one question: **is this project actually ready to ship?**

It checks the small things that break releases: uncommitted changes, failing tests, a broken production build, missing environment variables, a missing README or license, a version number you already released, and missing CI. It then exits with a code your scripts can trust.

- **Local and offline.** No account, no API key, no network access, no telemetry.
- **Zero setup.** Detects test and build commands for Go, Node.js, Python, Java, PHP and Rust.
- **Read-only.** Never edits, deletes or commits your files.
- **Cross-platform.** One Go binary for Linux, macOS and Windows.

## Why it exists

You finished the feature. Tests pass. You deploy. Then something breaks because you forgot one small thing: a file you never committed, an API key missing in production, a build that only fails in production mode.

Larger teams catch these with release checklists and release managers. Solo developers and small teams usually don't have either. Shipcheck is that checklist, as one command.

## Installation

With Go 1.22 or newer:

```bash
go install github.com/App-Chef/shipcheck/cli/cmd/shipcheck@latest
```

Or build from source:

```bash
git clone https://github.com/App-Chef/shipcheck.git
cd shipcheck/cli
go build -o shipcheck ./cmd/shipcheck
```

Prebuilt binaries and package managers are not available yet.

## Quick start

```bash
cd your-project
shipcheck              # run every check
shipcheck --verbose    # show commands, output tails and timings
shipcheck init         # write a commented .shipcheck.yml
shipcheck && ./deploy.sh
```

| Command / flag      | Description                                           |
| ------------------- | ----------------------------------------------------- |
| `shipcheck`         | Run all checks (same as `shipcheck check`)            |
| `shipcheck init`    | Create `.shipcheck.yml` (`--force` to overwrite)      |
| `shipcheck version` | Print the version (`--verbose` for build details)     |
| `shipcheck help`    | Help for any command                                  |
| `--verbose`         | Commands, output excerpts and timings                 |
| `--json`            | JSON report on stdout                                 |
| `-c, --config`      | Use a specific config file                            |
| `-C, --dir`         | Check another directory                               |
| `--no-color`        | Disable color (`NO_COLOR` is honoured too)            |
| `--version`         | Print the version                                     |

## Checks

| ID            | Checks                                                                                 | Default result when it finds a problem |
| ------------- | -------------------------------------------------------------------------------------- | -------------------------------------- |
| `git.repo`    | The project is in a Git repository                                                     | fail                                   |
| `git.clean`   | No uncommitted or untracked changes                                                    | fail                                   |
| `git.branch`  | You're on a branch you ship from (`main`/`master` or `git.protected_branches`)         | warn                                   |
| `tests`       | The detected (or configured) test command passes                                       | fail                                   |
| `build`       | The detected (or configured) production build passes                                  | fail                                   |
| `environment` | Variables from `.env.example` and `environment.required` are set (values never shown) | fail (warn for a committed `.env`)     |
| `readme`      | A non-empty `README.md` exists                                                         | warn                                   |
| `license`     | `LICENSE`, `LICENSE.md` or `LICENSE.txt` exists                                        | warn                                   |
| `version`     | A version is declared, and isn't already tagged at an older commit                     | warn                                   |
| `ci`          | GitHub Actions, GitLab CI, CircleCI, Jenkins, Bitbucket or Azure Pipelines config      | warn                                   |

Detected commands:

| Ecosystem | Tests                                        | Build                                   |
| --------- | -------------------------------------------- | --------------------------------------- |
| Go        | `go test ./...`                              | `go build ./...`                        |
| Node.js   | `npm test` (or pnpm / yarn / bun)            | `npm run build` if a build script exists |
| Python    | `pytest` / `python -m pytest`                | set `commands.build`                    |
| Java      | `mvn test` / `gradle test` (wrappers first)  | `mvn package -DskipTests` / `gradle assemble` |
| PHP       | `php artisan test`, `composer test`, PHPUnit | set `commands.build`                    |
| Rust      | `cargo test`                                 | `cargo build --release`                 |

Details for every check: [docs/checks.md](docs/checks.md).

## Configuration

Shipcheck works without configuration. To change something, add `.shipcheck.yml` (or run `shipcheck init`):

```yaml
version: 1

checks:
  license: false        # turn a check off

severity:
  ci: error             # warnings → failures
  git_clean: warn       # failures → warnings

git:
  protected_branches: [main, release]

environment:
  required: [DATABASE_URL, API_KEY]
  files: [.env, .env.local]

commands:
  tests: npm run test:ci
  build:
    - npm run build
  timeout: 10m
```

Unknown keys, invalid values and shell syntax in commands are rejected with exit code 2. Commands run **without a shell**. Full reference: [docs/configuration.md](docs/configuration.md).

## CI

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: stable
- run: go install github.com/App-Chef/shipcheck/cli/cmd/shipcheck@latest
- run: shipcheck --verbose
```

A failed check fails the job. See [docs/ci.md](docs/ci.md) for GitHub Actions and GitLab CI examples. This repository checks itself with its own [`.shipcheck.yml`](.shipcheck.yml) in [CI](.github/workflows/ci.yml).

## JSON output

```bash
shipcheck --json
```

```json
{
  "ready": false,
  "version": "v0.1.0",
  "summary": { "passed": 6, "warnings": 1, "failed": 3, "skipped": 0 },
  "duration_ms": 352,
  "checks": [
    {
      "id": "git.clean",
      "name": "Working tree",
      "description": "Checks that there are no uncommitted or untracked changes.",
      "status": "fail",
      "message": "1 uncommitted change",
      "suggestion": "Commit or stash your changes so you ship exactly what is in Git.",
      "duration_ms": 9
    }
  ]
}
```

Only JSON is written to stdout. Errors are reported as `{"error": "..."}`. Format reference: [docs/json-output.md](docs/json-output.md).

## Exit codes

| Code | Meaning                                                      |
| ---- | ------------------------------------------------------------ |
| `0`  | Ready: nothing failed (warnings allowed)                     |
| `1`  | Not ready: at least one check failed                         |
| `2`  | Shipcheck error: bad config, unknown flag, missing directory |

```bash
shipcheck && echo "Ready to deploy"
```

## Security

Shipcheck never makes network requests, uploads files, prints secret values, or modifies your source. It **does** run your project's own test and build commands, so only run it on code you trust. See [SECURITY.md](SECURITY.md).

## Development

```text
cli/   Go CLI: cmd/shipcheck, internal/{app,checks,config,detector,output,project,runner}, tests/
web/   Next.js website: docs and product site (static, no backend)
docs/  Markdown documentation
```

```bash
# CLI (Go 1.22+)
cd cli
go test ./...
go run ./cmd/shipcheck --dir .. --verbose

# Website (Node.js 20.9+)
cd web
npm install
npm run dev      # http://localhost:3000
npm run build
```

Set `NEXT_PUBLIC_SITE_URL` when building the website for deployment, so metadata and the sitemap use the real URL.

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md), including how to add a check, and follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE)
