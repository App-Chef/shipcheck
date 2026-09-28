import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2, Note, TableWrap } from "@/components/docs/doc-page";
import { allChecks } from "@/lib/checks";

export const metadata: Metadata = {
  title: "Configuration",
  description: "Configure Shipcheck with .shipcheck.yml: checks, severity, branches, environment and commands.",
};

const fullExample = `version: 1

checks:
  tests: true
  build: true
  environment: true
  readme: true
  license: false
  ci: false
  git_clean: true

severity:
  version: error   # a stale version blocks shipping
  readme: warn

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
  tests: npm run test:ci
  build:
    - npm run build
    - go build ./...
  timeout: 10m`;

export default function Configuration() {
  return (
    <DocPage
      href="/docs/configuration"
      eyebrow="Reference"
      title="Configuration"
      lead={
        <>
          Shipcheck works without configuration. When you need to change something, add a{" "}
          <code className="kbd-inline">.shipcheck.yml</code> to the project root.
        </>
      }
    >
      <p>
        Shipcheck looks for <code>.shipcheck.yml</code> (or <code>.shipcheck.yaml</code>) in the directory being
        checked. Use <code>--config path/to/file.yml</code> to load a different file. Run{" "}
        <code>shipcheck init</code> to generate a commented starting point.
      </p>
      <CodeBlock lang="yaml" title=".shipcheck.yml" code={fullExample} />
      <Note title="Strict by design">
        <p>
          Unknown keys, unknown check names, invalid severities and unsafe commands are rejected with an error and
          exit code 2. A typo never silently turns a check off.
        </p>
      </Note>

      <H2 id="version">version</H2>
      <p>
        The config format version. Currently <code>1</code>. Required so that future formats can change without
        breaking existing files.
      </p>

      <H2 id="checks">checks</H2>
      <p>
        Turn checks on or off. Every check is on by default, and disabled checks don&apos;t appear in the report.
        Keys use underscores; the dotted IDs from the JSON output (<code>git.clean</code>) are accepted too.
      </p>
      <TableWrap>
        <table className="doc-table">
          <thead>
            <tr>
              <th scope="col">Key</th>
              <th scope="col">Check</th>
            </tr>
          </thead>
          <tbody>
            {allChecks.map((c) => (
              <tr key={c.id}>
                <td><code>{c.key}</code></td>
                <td>
                  <Link href={`/docs/checks#${c.id.replace(".", "-")}`}>{c.name}</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </TableWrap>

      <H2 id="severity">severity</H2>
      <p>Change how a check&apos;s problems are reported, without turning it off.</p>
      <ul>
        <li>
          <code>error</code> turns warnings into failures. Use it for things you want to enforce, such as{" "}
          <code>ci: error</code>.
        </li>
        <li>
          <code>warn</code> turns failures into warnings. The problem is still shown, but doesn&apos;t block
          shipping, for example <code>git_clean: warn</code> while you iterate.
        </li>
      </ul>
      <p>
        Passing and skipped checks are unaffected. <code>warning</code> is accepted as an alias for <code>warn</code>.
      </p>
      <CodeBlock
        lang="yaml"
        code={`severity:
  ci: error        # require CI
  license: error   # require a license
  git_clean: warn  # allow uncommitted changes`}
      />

      <H2 id="git">git</H2>
      <p>
        <code>protected_branches</code> lists the branches you ship from. The branch check warns when you run
        Shipcheck on any other branch. Without this setting, Shipcheck uses <code>main</code> or <code>master</code>{" "}
        if either exists.
      </p>
      <CodeBlock
        lang="yaml"
        code={`git:
  protected_branches:
    - main
    - release`}
      />

      <H2 id="environment">environment</H2>
      <p>
        Variables listed in <code>.env.example</code>, <code>.env.sample</code> or <code>.example.env</code> are
        required automatically. Use <code>required</code> to add more. Each must have a non-empty value in the
        process environment or in one of <code>files</code> (default: <code>.env</code>).
      </p>
      <CodeBlock
        lang="yaml"
        code={`environment:
  required:
    - DATABASE_URL
    - STRIPE_SECRET_KEY
  files:
    - .env
    - .env.production`}
      />
      <p>
        Names must be valid variable names (<code>A-Z</code>, <code>0-9</code>, <code>_</code>). Files must be
        relative paths inside the project. Values are never printed.
      </p>

      <H2 id="commands">commands</H2>
      <p>
        Override the detected test and build commands. Each can be a single command or a list; commands in a list
        run in order and stop at the first failure. <code>timeout</code> applies to each command (default{" "}
        <code>10m</code>).
      </p>
      <CodeBlock
        lang="yaml"
        code={`commands:
  tests:
    - go -C api test ./...
    - npm --prefix web test
  build: make release
  timeout: 15m`}
      />
      <Note title="Commands run without a shell">
        <p>
          Shipcheck splits commands into arguments itself and runs them directly. Quotes work, but pipes, redirects,{" "}
          <code>&amp;&amp;</code>, <code>;</code>, <code>$VARIABLES</code> and globs are rejected. Put anything more
          complex in a script, npm script or Makefile target and run that. See the{" "}
          <Link href="/docs/security">security model</Link>.
        </p>
      </Note>
    </DocPage>
  );
}
