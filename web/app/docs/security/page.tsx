import type { Metadata } from "next";
import { DocPage, H2, Note } from "@/components/docs/doc-page";
import { repoFile } from "@/lib/site";

export const metadata: Metadata = {
  title: "Security model",
  description: "What Shipcheck runs on your machine, what it reads, and what it never does.",
};

export default function Security() {
  return (
    <DocPage
      href="/docs/security"
      eyebrow="Project"
      title="Security model"
      lead="Shipcheck runs on your machine, reads your project and runs your own commands. Here is exactly what that means."
    >
      <H2 id="never">What Shipcheck never does</H2>
      <ul>
        <li>Make network requests. There is no telemetry, update check or analytics.</li>
        <li>Upload, copy or send project files anywhere.</li>
        <li>Print the values of environment variables or secrets. Only variable names are shown.</li>
        <li>Modify, delete or commit your source files.</li>
        <li>Run commands through a shell.</li>
      </ul>

      <H2 id="reads">What it reads</H2>
      <ul>
        <li>Project manifests (<code>package.json</code>, <code>go.mod</code>, <code>Cargo.toml</code>, …) to detect commands and the version.</li>
        <li>Git metadata, through read-only commands: <code>status</code>, <code>rev-parse</code>, <code>symbolic-ref</code>, <code>ls-files</code> and <code>describe</code>.</li>
        <li>Env templates and the dotenv files you configure, to check that variables have values. Values stay in memory and are only used to redact them from output.</li>
        <li>The presence of README, LICENSE and CI files.</li>
      </ul>

      <H2 id="runs">What it runs</H2>
      <p>
        The tests and build checks run your project&apos;s own commands, such as <code>npm test</code> or{" "}
        <code>cargo build --release</code>. These do whatever your scripts do, just as if you ran them yourself.
        Builds write their normal output (such as <code>dist/</code>).
      </p>
      <ul>
        <li>
          Commands are split into arguments and executed directly, without a shell. Pipes, redirects, command
          chaining, variable expansion and globs in <code>.shipcheck.yml</code> are rejected, not interpreted.
        </li>
        <li>Commands run in the project directory, with no standard input, and are stopped after a timeout (10 minutes by default).</li>
        <li>
          Output is captured, not streamed. Only the last 20 lines of a failing command are shown, and only with{" "}
          <code>--verbose</code>. Values of required variables are replaced with <code>****</code>.
        </li>
        <li>Config values are validated: env file paths must stay inside the project, and variable names must be valid.</li>
      </ul>

      <Note title="Untrusted projects">
        <p>
          Running Shipcheck in a project runs that project&apos;s test and build scripts. Treat it like running{" "}
          <code>npm test</code>: only run it on code you trust, or turn off the <code>tests</code> and{" "}
          <code>build</code> checks.
        </p>
      </Note>

      <H2 id="reporting">Reporting a vulnerability</H2>
      <p>
        Please don&apos;t open a public issue for security problems. Follow the process in{" "}
        <a href={repoFile("SECURITY.md")}>SECURITY.md</a>.
      </p>
    </DocPage>
  );
}
