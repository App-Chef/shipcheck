# JSON output and exit codes

## `--json`

`shipcheck --json` prints one JSON document to stdout and nothing else: no progress, colors or text.

```json
{
  "ready": false,
  "version": "v0.1.0",
  "summary": { "passed": 6, "warnings": 1, "failed": 3, "skipped": 0 },
  "duration_ms": 352,
  "checks": [
    {
      "id": "tests",
      "name": "Tests",
      "description": "Detects and runs the project's test command (go test, npm test, pytest, cargo test, …).",
      "status": "fail",
      "message": "Tests failed (go test ./...)",
      "suggestion": "Run `go test ./...` to see the full output.",
      "duration_ms": 298
    }
  ]
}
```

| Field         | Type    | Description                                            |
| ------------- | ------- | ------------------------------------------------------ |
| `ready`       | boolean | `true` when no check failed                            |
| `version`     | string  | Shipcheck version                                      |
| `summary`     | object  | `passed`, `warnings`, `failed`, `skipped`              |
| `duration_ms` | number  | Total run time                                         |
| `checks`      | array   | One entry per enabled check, in run order              |

Each check has `id` (stable, e.g. `git.clean`), `name`, `description`, `status` (`pass`, `warn`, `fail`, `skip`, after severity), `message`, optional `suggestion`, `duration_ms`, and, with `--verbose` only, `details` (commands run, changed files, output tails with secrets redacted).

If Shipcheck itself fails, it prints `{"error": "..."}` and exits with code 2.

## Exit codes

| Code | Meaning                                                                         |
| ---- | ------------------------------------------------------------------------------- |
| `0`  | Ready. No check failed; warnings and skips are allowed.                         |
| `1`  | Not ready. At least one check failed.                                           |
| `2`  | Shipcheck error: invalid config, unknown flag, missing directory, interrupted.  |

```bash
shipcheck && echo "Ready to deploy"
```
