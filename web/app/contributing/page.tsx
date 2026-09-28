import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { H2, Note } from "@/components/docs/doc-page";
import { repoFile, site } from "@/lib/site";

export const metadata: Metadata = {
  title: "Contributing",
  description: "How to build, test and contribute to Shipcheck, including how to add a new check.",
};

const newCheck = `// cli/internal/checks/changelog.go
package checks

import "context"

// Changelog checks for a CHANGELOG file.
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
}`;

export default function Contributing() {
  return (
    <div className="container-page py-12 sm:py-16">
      <header className="max-w-3xl border-b border-subtle pb-8">
        <p className="eyebrow">Project</p>
        <h1 className="mt-3 text-[2.25rem] leading-[1.1] font-semibold tracking-[-0.03em] sm:text-5xl">
          Contributing
        </h1>
        <p className="mt-4 text-lg leading-relaxed text-muted">
          Shipcheck is small on purpose. The best contributions make existing checks more accurate, clearer or
          faster.
        </p>
      </header>

      <div className="prose-doc mt-8 max-w-[46rem]">
        <H2 id="layout">Repository layout</H2>
        <CodeBlock
          lang="text"
          title="shipcheck/"
          code={`cli/                 Go CLI (the tool itself)
  cmd/shipcheck/     entry point
  internal/app/      commands, flags, exit codes
  internal/checks/   one file per check
  internal/config/   .shipcheck.yml parsing and validation
  internal/detector/ test, build and version detection
  internal/output/   text and JSON renderers
  internal/project/  filesystem and Git helpers
  internal/runner/   runs checks, applies severity
  tests/             end-to-end tests against the real binary
web/                 this website (Next.js)
docs/                documentation in Markdown`}
        />

        <H2 id="cli">Working on the CLI</H2>
        <p>You need Go 1.22 or newer and Git.</p>
        <CodeBlock
          lang="bash"
          code={`$ cd cli
$ go test ./...
$ go run ./cmd/shipcheck --dir .. --verbose`}
        />
        <p>
          Tests create real temporary projects and Git repositories, so they need no fixtures and no network. Please
          add a test for every behaviour change.
        </p>

        <H2 id="new-check">Adding a check</H2>
        <p>
          A check is a value with an ID, a name, a description and a <code>Run</code> function. The runner handles
          timing, severity, enabling and output, so a check only decides its result.
        </p>
        <CodeBlock lang="go" title="cli/internal/checks/changelog.go" code={newCheck} />
        <ol>
          <li>
            Add it to <code>All()</code> in <code>cli/internal/checks/checks.go</code>, in the position it should run.
          </li>
          <li>
            Add tests next to it. The existing tests show how to build a throwaway project with{" "}
            <code>testutil.Dir</code>.
          </li>
          <li>
            Document it in <code>web/lib/checks.ts</code> and <code>docs/checks.md</code>.
          </li>
        </ol>
        <Note title="Guidelines for checks">
          <ul className="list-[square] pl-5">
            <li>Fast and offline. No network access, ever.</li>
            <li>Read-only. Never modify the project.</li>
            <li>Never print secret values.</li>
            <li>Short messages that describe the result (“No changelog”), with an actionable suggestion.</li>
            <li>Fail only for real release blockers; warn for everything else.</li>
          </ul>
        </Note>

        <H2 id="web">Working on the website</H2>
        <p>You need Node.js 20.9 or newer.</p>
        <CodeBlock
          lang="bash"
          code={`$ cd web
$ npm install
$ npm run dev
$ npm run lint && npm run build`}
        />
        <p>
          The site is static: no database, no API routes, no analytics. Terminal examples are copied from real CLI
          output, so update them when CLI messages change.
        </p>

        <H2 id="pull-requests">Pull requests</H2>
        <ul>
          <li>Open an issue first for new checks or larger changes, so we can agree on scope.</li>
          <li>Keep pull requests focused on one change.</li>
          <li>
            Run <code>gofmt</code>, <code>go vet ./...</code> and <code>go test ./...</code> for CLI changes, and{" "}
            <code>npm run lint</code> and <code>npm run build</code> for website changes.
          </li>
          <li>
            Follow the <a href={repoFile("CODE_OF_CONDUCT.md")}>code of conduct</a>.
          </li>
        </ul>
        <p>
          The full guide is in <a href={repoFile("CONTRIBUTING.md")}>CONTRIBUTING.md</a>. Issues and pull requests
          live on <a href={site.repo}>GitHub</a>. New to Shipcheck? Start with the{" "}
          <Link href="/docs">documentation</Link>.
        </p>
      </div>
    </div>
  );
}
