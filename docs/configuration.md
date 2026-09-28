# Configuration

Shipcheck works without configuration. To change its behaviour, add `.shipcheck.yml` (or `.shipcheck.yaml`) to the directory being checked, or pass `--config path`. `shipcheck init` writes a commented starting point based on your project; it won't overwrite an existing file without `--force`.

```yaml
version: 1

checks:
  license: false
  ci: true

severity:
  ci: error
  git_clean: warn

git:
  protected_branches:
    - main
    - master

environment:
  required:
    - DATABASE_URL
    - API_KEY
  files:
    - .env
    - .env.local

commands:
  tests: go test ./...
  build:
    - npm run build
    - go build ./...
  timeout: 10m
```

The config is strict: unknown keys, unknown check names, invalid severities, invalid variable names, unsafe paths and shell syntax in commands are errors (exit code 2). A typo never silently disables a check.

## `version`

Config format version. Must be `1`.

## `checks`

Enable (`true`) or disable (`false`) checks. All checks are enabled by default; disabled checks don't appear in the report.

Keys: `git_repo`, `git_clean`, `git_branch`, `tests`, `build`, `environment`, `readme`, `license`, `version`, `ci`. Dotted IDs (`git.clean`) are accepted too.

## `severity`

Changes how a check's problems are reported:

- `error`: warnings become failures (block shipping)
- `warn` (or `warning`): failures become warnings (reported, don't block)

Passing and skipped checks are unaffected.

## `git.protected_branches`

The branches you ship from. The branch check warns on any other branch. Default: `main` or `master`, whichever exists.

## `environment`

- `required`: variable names that must have a non-empty value, in addition to those listed in `.env.example`, `.env.sample` or `.example.env`.
- `files`: dotenv files that may provide values, relative to the project. Default: `[.env]`. The process environment takes precedence.

Values are never printed.

## `commands`

- `tests`, `build`: a command string or a list of commands, replacing detection. Lists run in order and stop at the first failure.
- `timeout`: per-command timeout, such as `90s` or `15m`. Default `10m`.

Commands are split into arguments by Shipcheck and executed **without a shell**. Single and double quotes work. Pipes, redirects, `&&`, `;`, `$VARS`, `$(...)`, backticks and globs are rejected; put complex steps in a script, npm script or Makefile target instead. For monorepos, use tools' own directory flags:

```yaml
commands:
  tests:
    - go -C api test ./...
    - npm --prefix web test
```
