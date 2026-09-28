# Checks

Shipcheck runs these checks in order. Each one passes (`✓`), warns (`⚠`), fails (`✗`) or is skipped (`–`). Only failures block shipping (exit code 1). Every check can be disabled under `checks:` or have its [severity](configuration.md#severity) changed. Config keys are the IDs with `.` replaced by `_`.

## Git repository (`git.repo`)

Checks that the project is inside a Git repository (from any subdirectory, so monorepos work).

- Why: a release that isn't in version control can't be traced, reviewed or rolled back.
- Results: `✓ Git repository` · `✗ Not a Git repository`

## Working tree (`git.clean`)

Fails when `git status` reports uncommitted or untracked changes. Ignored files don't count. `--verbose` lists the first ten paths.

- Why: if what you deploy differs from what's committed, the release can't be reproduced.
- Results: `✓ Working tree clean` · `✗ 3 uncommitted changes` · `– Working tree not checked (not a Git repository)`

## Branch (`git.branch`)

Warns when you aren't on a branch you ship from: `git.protected_branches`, or `main`/`master` when they exist. Skipped on a detached HEAD (common in CI) or when no release branch is known.

- Why: shipping from a feature branch releases code that was never merged.
- Results: `✓ On branch main` · `⚠ On branch feature/login, not main` · `– Detached HEAD (no branch to check)`

## Tests (`tests`)

Detects and runs the test command with `CI=true` (so watch-mode runners exit). Output is captured. On failure, `--verbose` shows the last 20 lines with secrets redacted.

| Ecosystem | Detected by                                                  | Command                                              |
| --------- | ------------------------------------------------------------ | ---------------------------------------------------- |
| Go        | `go.mod`                                                     | `go test ./...`                                      |
| Node.js   | `test` script in `package.json` (not the npm placeholder)    | `npm test`, `pnpm run test`, `yarn run test`, `bun run test` (from `packageManager` or lockfile) |
| Python    | `pytest.ini`, `conftest.py`, `tests/`, `[tool.pytest…]`      | `pytest` or `python -m pytest`                       |
| Java      | `pom.xml` / `build.gradle(.kts)`                             | `./mvnw test`, `mvn test`, `./gradlew test`, `gradle test` |
| PHP       | `artisan`, `composer.json` `test` script, PHPUnit config     | `php artisan test`, `composer test`, `vendor/bin/phpunit` |
| Rust      | `Cargo.toml`                                                 | `cargo test`                                         |

- Why: tests only protect a release if you actually ran them on the code you ship.
- Results: `✓ Tests passed` · `✗ Tests failed (npm test)` · `✗ Tests not run: cargo is not installed` · `– No test command detected`
- Override: `commands.tests` in `.shipcheck.yml`.

## Production build (`build`)

Detects and runs the production build: `go build ./...`, the `build` script in `package.json`, `mvn package -DskipTests`, `gradle assemble`, or `cargo build --release`. Python and PHP have no standard build; set `commands.build` if you have one.

- Why: many errors only appear in production mode.
- Results: `✓ Production build passed` · `✗ Production build failed (npm run build)` · `✗ npm run build timed out after 10m0s` · `– No build command detected`

## Environment (`environment`)

Reads variable names from `.env.example`, `.env.sample` or `.example.env`, plus `environment.required`. Each must have a non-empty value in the process environment or in `environment.files` (default `.env`). Also warns when a real `.env` file is tracked by Git. **Values are never printed**, and are redacted from any command output Shipcheck shows.

- Why: a missing API key is the classic "works locally, breaks in production" bug.
- Results: `✓ Environment checked` · `✗ 2 environment variables missing` · `⚠ .env is committed to Git` · `– No environment variables required`

## README (`readme`)

Looks for `README.md` (case-insensitive; `README`, `README.markdown`, `README.rst` and `README.txt` also count). An empty README warns.

- Results: `✓ README found` · `⚠ No README` · `⚠ README.md is empty`

## License (`license`)

Looks for `LICENSE`, `LICENSE.md` or `LICENSE.txt` (also `LICENCE`, `COPYING`) and names common licenses (MIT, Apache-2.0, GPL, AGPL, LGPL, MPL-2.0, ISC, BSD, Unlicense). Private projects can disable it.

- Results: `✓ License found (MIT)` · `⚠ No license file`

## Version (`version`)

Reads the version from `package.json`, `Cargo.toml`, `pyproject.toml`, `composer.json`, `pom.xml`, `build.gradle(.kts)` or `VERSION`, falling back to the latest Git tag. Warns if a tag for that version (`1.4.0` or `v1.4.0`) exists at an older commit than `HEAD`: the version was already released.

- Results: `✓ Version 1.4.0` · `⚠ Version 1.4.0 was already released (v1.4.0)` · `⚠ No version found`

## CI (`ci`)

Detects GitHub Actions (`.github/workflows/*.yml`), GitLab CI, CircleCI, Jenkins, Bitbucket Pipelines and Azure Pipelines. Missing CI is a warning unless you set `severity: { ci: error }`.

- Results: `✓ CI configured (GitHub Actions)` · `⚠ No CI configuration`
