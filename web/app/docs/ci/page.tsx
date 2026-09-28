import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2, Note } from "@/components/docs/doc-page";
import { site } from "@/lib/site";

export const metadata: Metadata = {
  title: "CI",
  description: "Run Shipcheck in GitHub Actions, GitLab CI or any CI system, and gate deploys on the result.",
};

const githubActions = `name: Shipcheck

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
          fetch-depth: 0   # full history and tags for the version check

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      # Set up your own toolchain too, e.g. actions/setup-node + npm ci.

      - name: Install Shipcheck
        run: ${site.installCommand}

      - name: Run Shipcheck
        run: shipcheck --verbose
        env:
          DATABASE_URL: \${{ secrets.DATABASE_URL }}`;

const gitlab = `shipcheck:
  image: golang:1.24
  variables:
    GIT_DEPTH: 0
  script:
    - ${site.installCommand}
    - shipcheck --verbose`;

const jsonStep = `      - name: Run Shipcheck
        run: shipcheck --json > shipcheck.json

      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: shipcheck-report
          path: shipcheck.json`;

export default function CI() {
  return (
    <DocPage
      href="/docs/ci"
      eyebrow="Reference"
      title="CI"
      lead="Shipcheck runs the same way in CI as on your laptop. A failed check exits with code 1, which fails the job."
    >
      <H2 id="github-actions">GitHub Actions</H2>
      <p>
        Install Shipcheck with Go, set up whatever your tests and build need, and run it. Pass required environment
        variables from repository secrets. Shipcheck checks that they&apos;re set, never what they contain.
      </p>
      <CodeBlock lang="yaml" title=".github/workflows/shipcheck.yml" code={githubActions} />
      <Note title="Checkout depth">
        <p>
          <code>actions/checkout</code> fetches a single commit by default, which hides tags from the version check.
          Use <code>fetch-depth: 0</code> if you rely on Git tags for versioning.
        </p>
      </Note>

      <H2 id="gitlab">GitLab CI</H2>
      <CodeBlock lang="yaml" title=".gitlab-ci.yml" code={gitlab} />

      <H2 id="other">Any other CI</H2>
      <p>Shipcheck only needs the binary and your toolchain. The pattern is always the same:</p>
      <CodeBlock
        lang="bash"
        code={`$ ${site.installCommand}
$ shipcheck --verbose`}
      />
      <p>
        <code>--verbose</code> is useful in CI logs: it shows which commands ran, how long they took, and the tail
        of any failing output.
      </p>

      <H2 id="json-reports">Keep a JSON report</H2>
      <p>
        <code>--json</code> prints a machine-readable report and keeps the same exit codes. Save it as a build
        artifact, or parse it to post a summary. See <Link href="/docs/json-output">JSON output</Link> for the
        format.
      </p>
      <CodeBlock lang="yaml" title="GitHub Actions steps" code={jsonStep} />

      <H2 id="gate-deploys">Gate a deploy</H2>
      <p>Because the exit code is 0 only when nothing failed, Shipcheck composes with any deploy command:</p>
      <CodeBlock
        lang="bash"
        code={`$ shipcheck && ./deploy.sh
$ shipcheck --json | jq -e '.ready' && npm run deploy`}
      />

      <H2 id="ci-differences">Differences in CI</H2>
      <ul>
        <li>
          <strong>Branch.</strong> CI systems often check out a detached HEAD, so the branch check is skipped rather
          than failed.
        </li>
        <li>
          <strong>Working tree.</strong> Steps that write files (such as a build output that isn&apos;t in{" "}
          <code>.gitignore</code>) make the tree dirty. Ignore those paths, or set <code>git_clean: warn</code>.
        </li>
        <li>
          <strong>Environment.</strong> There is usually no <code>.env</code> file in CI. Provide required variables
          through your CI&apos;s secret settings.
        </li>
      </ul>
      <p>
        Shipcheck&apos;s own repository runs itself in CI. See{" "}
        <a href={`${site.repo}/blob/main/.github/workflows/ci.yml`}>ci.yml</a>.
      </p>
    </DocPage>
  );
}
