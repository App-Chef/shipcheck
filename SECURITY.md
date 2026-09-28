# Security policy

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Report vulnerabilities privately through GitHub: go to the repository's **Security** tab and choose **Report a vulnerability**. Include the Shipcheck version (`shipcheck version --verbose`), your operating system, and steps to reproduce.

You should get a response within a week. Once a fix is ready, we'll publish a release and a security advisory, crediting you unless you'd rather stay anonymous.

## Supported versions

Security fixes are made on the latest release only.

## Security model

Shipcheck runs locally and treats your project as private.

**Shipcheck never:**

- makes network requests (no telemetry, update checks or analytics)
- uploads or copies project files
- prints the values of environment variables or secrets (only names)
- modifies, deletes or commits your files
- runs commands through a shell

**What it does:**

- Reads manifests, Git metadata (read-only `git` commands), env templates, configured dotenv files, and the presence of README, LICENSE and CI files.
- Runs your project's own test and build commands (detected or from `.shipcheck.yml`). They run in the project directory with no stdin and a timeout (10 minutes by default). Builds write their usual output.
- Parses commands from `.shipcheck.yml` into arguments itself and executes them directly. Pipes, redirects, `&&`, `;`, variable expansion, command substitution and globs are **rejected**, not interpreted.
- Validates config: unknown keys are errors, env file paths must be relative and stay inside the project, and variable names must be valid.
- Captures command output instead of streaming it. With `--verbose`, the last 20 lines of a failing command are shown, with values of required variables replaced by `****`.

**Trust boundary:** running Shipcheck in a project runs that project's scripts, exactly like running `npm test` or `make`. Only run it on code you trust, or disable the `tests` and `build` checks:

```yaml
version: 1
checks:
  tests: false
  build: false
```
