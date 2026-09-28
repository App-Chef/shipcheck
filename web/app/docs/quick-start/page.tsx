import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2, Note, TableWrap } from "@/components/docs/doc-page";
import { Terminal } from "@/components/terminal";
import { notReadyTranscript, parseTranscript, verboseTranscript } from "@/lib/terminal";

export const metadata: Metadata = {
  title: "Quick start",
  description: "Run your first Shipcheck, read the results, and fix what matters.",
};

export default function QuickStart() {
  return (
    <DocPage
      href="/docs/quick-start"
      eyebrow="Getting started"
      title="Quick start"
      lead="Run one command in your project, read the report, fix what blocks you, then ship."
    >
      <H2 id="run">1. Run it</H2>
      <p>From the root of your project:</p>
      <CodeBlock lang="bash" code="$ shipcheck" />
      <p>
        <code>shipcheck</code> and <code>shipcheck check</code> are the same command. To check another directory,
        pass <code>--dir</code> (or <code>-C</code>):
      </p>
      <CodeBlock lang="bash" code="$ shipcheck --dir ./services/api" />

      <H2 id="read">2. Read the report</H2>
      <p>Each line is one check. When something needs attention, a suggestion follows it.</p>
      <Terminal lines={parseTranscript(notReadyTranscript)} label="Example report with failures" />
      <TableWrap>
        <table className="doc-table">
          <thead>
            <tr>
              <th scope="col">Symbol</th>
              <th scope="col">Status</th>
              <th scope="col">Meaning</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="font-mono text-[#1f8a43]">✓</td>
              <td><code>pass</code></td>
              <td>Nothing to do.</td>
            </tr>
            <tr>
              <td className="font-mono text-[#9a6b00]">⚠</td>
              <td><code>warn</code></td>
              <td>Worth fixing, but doesn&apos;t block shipping.</td>
            </tr>
            <tr>
              <td className="font-mono text-[#c2321f]">✗</td>
              <td><code>fail</code></td>
              <td>Blocks shipping. Shipcheck exits with code 1.</td>
            </tr>
            <tr>
              <td className="font-mono text-muted">–</td>
              <td><code>skip</code></td>
              <td>Not applicable, for example no test command found.</td>
            </tr>
          </tbody>
        </table>
      </TableWrap>

      <H2 id="dig-in">3. Dig into a failure</H2>
      <p>
        <code>--verbose</code> adds timings, the commands that ran, changed files and the last 20 lines of output
        from failing tests or builds. Secret values are redacted.
      </p>
      <Terminal lines={parseTranscript(verboseTranscript)} label="Verbose output excerpt" title="~/demo (excerpt)" />

      <H2 id="configure">4. Adjust it (optional)</H2>
      <p>The defaults work for most projects. To change them, generate a config file:</p>
      <CodeBlock
        lang="bash"
        code={`$ shipcheck init
✓ Created .shipcheck.yml
  Edit it to fit your project, then run \`shipcheck\`.`}
      />
      <p>
        <code>init</code> detects your branch, test and build commands, and <code>.env.example</code>, and writes a
        commented file. It never overwrites an existing config unless you pass <code>--force</code>. See{" "}
        <Link href="/docs/configuration">Configuration</Link> for every option.
      </p>

      <H2 id="ship">5. Ship</H2>
      <p>Use the exit code to gate a deploy:</p>
      <CodeBlock lang="bash" code={`$ shipcheck && ./deploy.sh`} />
      <Note title="Tip">
        <p>
          Run Shipcheck in CI as well, so every push gets the same checks. See <Link href="/docs/ci">CI</Link>.
        </p>
      </Note>

      <H2 id="commands">Commands and flags</H2>
      <TableWrap>
        <table className="doc-table">
          <thead>
            <tr>
              <th scope="col">Command</th>
              <th scope="col">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr><td><code>shipcheck</code></td><td>Run all checks (same as <code>shipcheck check</code>).</td></tr>
            <tr><td><code>shipcheck init</code></td><td>Create a <code>.shipcheck.yml</code>. <code>--force</code> overwrites.</td></tr>
            <tr><td><code>shipcheck version</code></td><td>Print the version. <code>--verbose</code> adds build details.</td></tr>
            <tr><td><code>shipcheck help</code></td><td>Show help for any command.</td></tr>
          </tbody>
        </table>
      </TableWrap>
      <TableWrap>
        <table className="doc-table">
          <thead>
            <tr>
              <th scope="col">Flag</th>
              <th scope="col">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr><td><code>--verbose</code></td><td>Show commands, output excerpts and timings.</td></tr>
            <tr><td><code>--json</code></td><td>Print a <Link href="/docs/json-output">JSON report</Link> instead of text.</td></tr>
            <tr><td><code>-c, --config</code></td><td>Use a specific config file instead of <code>.shipcheck.yml</code>.</td></tr>
            <tr><td><code>-C, --dir</code></td><td>Check another directory.</td></tr>
            <tr><td><code>--no-color</code></td><td>Disable color. <code>NO_COLOR</code> is honoured too.</td></tr>
            <tr><td><code>--version</code></td><td>Print the version and exit.</td></tr>
          </tbody>
        </table>
      </TableWrap>
    </DocPage>
  );
}
