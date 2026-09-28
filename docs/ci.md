# CI

Shipcheck behaves the same in CI as locally. A failed check exits with code 1, which fails the job.

## GitHub Actions

```yaml
name: Shipcheck

on:
  push:
    branches: [main]
  pull_request:

jobs:
  shipcheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0   # tags for the version check

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      # Set up your own toolchain too, e.g. actions/setup-node + npm ci.

      - name: Install Shipcheck
        run: go install github.com/App-Chef/shipcheck/cli/cmd/shipcheck@latest

      - name: Run Shipcheck
        run: shipcheck --verbose
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

## GitLab CI

```yaml
shipcheck:
  image: golang:1.24
  variables:
    GIT_DEPTH: 0
  script:
    - go install github.com/App-Chef/shipcheck/cli/cmd/shipcheck@latest
    - shipcheck --verbose
```

## Keeping a report

```yaml
      - run: shipcheck --json > shipcheck.json
      - if: always()
        uses: actions/upload-artifact@v4
        with:
          name: shipcheck-report
          path: shipcheck.json
```

## Gating a deploy

```bash
shipcheck && ./deploy.sh
shipcheck --json | jq -e '.ready' && npm run deploy
```

## Things that differ in CI

- **Detached HEAD.** Many CI systems check out a commit, not a branch; the branch check is then skipped.
- **Build output.** Files written during CI that aren't in `.gitignore` make the working tree dirty. Ignore them, or set `severity: { git_clean: warn }`.
- **Secrets.** There's usually no `.env` in CI. Provide required variables as CI secrets.

This repository runs Shipcheck on itself: see [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) and [`.shipcheck.yml`](../.shipcheck.yml).
