import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2 } from "@/components/docs/doc-page";
import { checkGroups } from "@/lib/checks";
import { site } from "@/lib/site";

export const metadata: Metadata = {
  title: "Documentation",
  description: "Learn what Shipcheck checks, how to install it, and how to use it locally and in CI.",
};

export default function DocsIntroduction() {
  return (
    <DocPage
      href="/docs"
      eyebrow="Getting started"
      title="Introduction"
      lead="Shipcheck is a small command-line tool that answers one question: is this project actually ready to ship?"
    >
      <p>
        Run <code>shipcheck</code> in a project and it checks for the common release blockers: uncommitted changes,
        failing tests, a broken production build, missing environment variables, a missing README or license, an
        already-released version number and missing CI. Each check passes, warns, fails or is skipped, and the
        exit code tells scripts whether it&apos;s safe to continue.
      </p>
      <CodeBlock
        lang="bash"
        code={`$ shipcheck && ./deploy.sh`}
        title="Deploy only when ready"
      />

      <H2 id="principles">Principles</H2>
      <ul>
        <li>
          <strong>Local and offline.</strong> No account, no API key, no network access and no telemetry. Your code
          never leaves your machine.
        </li>
        <li>
          <strong>Zero setup.</strong> Test and build commands are detected for Go, Node.js, Python, Java, PHP and
          Rust. A config file is optional.
        </li>
        <li>
          <strong>Read-only.</strong> Shipcheck never edits, deletes or commits your files. The only files written
          are the ones your own build writes, plus <code>.shipcheck.yml</code> when you run{" "}
          <code>shipcheck init</code>.
        </li>
        <li>
          <strong>Scriptable.</strong> Stable exit codes and a <Link href="/docs/json-output">JSON report</Link> make
          it easy to use in CI and deploy scripts.
        </li>
      </ul>

      <H2 id="what-it-checks">What it checks</H2>
      <ul>
        {checkGroups.map((g) => (
          <li key={g.slug}>
            <Link href={`/docs/checks#${g.slug}`}>{g.title}</Link>: {g.summary}
          </li>
        ))}
      </ul>

      <H2 id="what-it-is-not">What it is not</H2>
      <p>
        Shipcheck is a pre-shipping validation tool, not a deployment tool. It doesn&apos;t deploy, monitor, lint
        your code or manage infrastructure. It runs the commands you already have and checks the things that are
        easy to forget.
      </p>

      <H2 id="next-steps">Next steps</H2>
      <ol>
        <li>
          <Link href="/docs/installation">Install Shipcheck</Link> (Go {site.minGo} or newer).
        </li>
        <li>
          Follow the <Link href="/docs/quick-start">quick start</Link> to run your first check.
        </li>
        <li>
          Tune it with <Link href="/docs/configuration">.shipcheck.yml</Link> if the defaults don&apos;t fit.
        </li>
      </ol>
    </DocPage>
  );
}
