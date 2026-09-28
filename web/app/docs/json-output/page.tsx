import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2, Note, TableWrap } from "@/components/docs/doc-page";

export const metadata: Metadata = {
  title: "JSON output",
  description: "The machine-readable report printed by shipcheck --json.",
};

const example = `{
  "ready": false,
  "version": "v0.1.0",
  "summary": {
    "passed": 6,
    "warnings": 1,
    "failed": 3,
    "skipped": 0
  },
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
    },
    {
      "id": "tests",
      "name": "Tests",
      "description": "Detects and runs the project's test command (go test, npm test, pytest, cargo test, …).",
      "status": "fail",
      "message": "Tests failed (go test ./...)",
      "suggestion": "Run \`go test ./...\` to see the full output.",
      "duration_ms": 298
    },
    {
      "id": "ci",
      "name": "CI",
      "description": "Detects CI configuration for GitHub Actions, GitLab CI, CircleCI, Jenkins, Bitbucket or Azure Pipelines.",
      "status": "warn",
      "message": "No CI configuration",
      "suggestion": "Add a CI workflow so every push is tested (shipcheck runs in CI too).",
      "duration_ms": 0
    }
  ]
}`;

type Field = [name: string, type: string, description: string];

const reportFields: Field[] = [
  ["ready", "boolean", "true when no check failed. Warnings don't affect it."],
  ["version", "string", "The Shipcheck version that produced the report."],
  ["summary", "object", "Counts: passed, warnings, failed, skipped."],
  ["duration_ms", "number", "Total run time in milliseconds."],
  ["checks", "array", "One entry per enabled check, in run order."],
];

const checkFields: Field[] = [
  ["id", "string", "Stable check ID, e.g. git.clean. Safe to match on."],
  ["name", "string", "Human-readable check name."],
  ["description", "string", "What the check does."],
  ["status", "string", "One of pass, warn, fail, skip (after severity is applied)."],
  ["message", "string", "The one-line result, as shown in the terminal."],
  ["suggestion", "string?", "How to fix it. Omitted when there is nothing to suggest."],
  ["duration_ms", "number", "How long the check took."],
  ["details", "string[]?", "Only with --verbose: commands run, changed files, output tails."],
];

function FieldTable({ fields, label }: { fields: Field[]; label: string }) {
  return (
    <TableWrap>
      <table className="doc-table">
        <caption className="sr-only">{label}</caption>
        <thead>
          <tr>
            <th scope="col">Field</th>
            <th scope="col">Type</th>
            <th scope="col">Description</th>
          </tr>
        </thead>
        <tbody>
          {fields.map(([name, type, desc]) => (
            <tr key={name}>
              <td><code>{name}</code></td>
              <td className="font-mono text-[13px] whitespace-nowrap text-muted">{type}</td>
              <td>{desc}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </TableWrap>
  );
}

export default function JSONOutput() {
  return (
    <DocPage
      href="/docs/json-output"
      eyebrow="Reference"
      title="JSON output"
      lead={
        <>
          <code className="kbd-inline">shipcheck --json</code> prints a single JSON document to stdout and nothing
          else, so it can be piped straight into other tools.
        </>
      }
    >
      <CodeBlock lang="bash" code="$ shipcheck --json" />
      <CodeBlock lang="json" title="stdout (abridged to three checks)" code={example} />

      <H2 id="report">Report fields</H2>
      <FieldTable fields={reportFields} label="Report fields" />

      <H2 id="check">Check fields</H2>
      <FieldTable fields={checkFields} label="Check fields" />

      <H2 id="errors">Errors</H2>
      <p>
        If Shipcheck itself can&apos;t run (for example an invalid config file), it still prints JSON, with a single{" "}
        <code>error</code> field, and exits with code 2:
      </p>
      <CodeBlock
        lang="json"
        code={`{
  "error": ".shipcheck.yml: line 3: invalid severity \\"loud\\" (use error or warn)"
}`}
      />

      <H2 id="guarantees">Guarantees</H2>
      <ul>
        <li>No progress output, colors or text is mixed into stdout.</li>
        <li>
          Exit codes are the same as in text mode. See <Link href="/docs/exit-codes">Exit codes</Link>.
        </li>
        <li>Secret values never appear. Only variable names do.</li>
        <li>Check IDs are stable. New fields may be added, but existing ones won&apos;t change meaning.</li>
      </ul>

      <H2 id="examples">Examples with jq</H2>
      <CodeBlock
        lang="bash"
        code={`# Is it ready?
$ shipcheck --json | jq '.ready'

# Only the problems
$ shipcheck --json | jq '.checks[] | select(.status == "fail" or .status == "warn") | {id, message}'

# Fail a script unless ready
$ shipcheck --json | jq -e '.ready' > /dev/null`}
      />
      <Note title="Details">
        <p>
          <code>details</code> is only included with <code>--verbose</code>, because it can contain test output.
          Values of required environment variables are redacted from it.
        </p>
      </Note>
    </DocPage>
  );
}
