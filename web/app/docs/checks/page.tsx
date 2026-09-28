import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2, H3 } from "@/components/docs/doc-page";
import { Terminal } from "@/components/terminal";
import { checkGroups, type CheckDoc } from "@/lib/checks";
import type { Line } from "@/lib/terminal";

export const metadata: Metadata = {
  title: "Checks",
  description: "Every check Shipcheck runs: what it looks for, why it exists, example output and configuration.",
};

function outcomeLines(check: CheckDoc): Line[] {
  const symbols = { pass: "✓", warn: "⚠", fail: "✗", skip: "–" } as const;
  return check.outcomes.map((o) => ({ tone: o.status, text: `${symbols[o.status]} ${o.message}` }));
}

function CheckBody({ check }: { check: CheckDoc }) {
  return (
    <>
      <dl className="flex flex-wrap gap-2 font-mono text-xs">
        <div className="flex items-center gap-1.5 rounded-[4px] border border-subtle bg-card px-2 py-1">
          <dt className="text-muted">id</dt>
          <dd>{check.id}</dd>
        </div>
        <div className="flex items-center gap-1.5 rounded-[4px] border border-subtle bg-card px-2 py-1">
          <dt className="text-muted">config key</dt>
          <dd>{check.key}</dd>
        </div>
      </dl>
      <p>
        <strong>What it checks.</strong> {check.what}
      </p>
      <p>
        <strong>Why it exists.</strong> {check.why}
      </p>
      {check.notes && (
        <ul>
          {check.notes.map((n) => (
            <li key={n}>{n}</li>
          ))}
        </ul>
      )}
      <p className="eyebrow !mt-6">Possible results</p>
      <Terminal
        lines={outcomeLines(check)}
        title={`${check.name.toLowerCase()} · possible results`}
        label={`Possible results for the ${check.name} check`}
        className="!shadow-hard"
      />
      <p className="eyebrow !mt-6">Configure</p>
      <CodeBlock
        lang="yaml"
        title=".shipcheck.yml"
        code={
          check.config ??
          `checks:
  ${check.key}: false   # turn it off

severity:
  ${check.key}: warn    # or error`
        }
      />
    </>
  );
}

export default function Checks() {
  return (
    <DocPage
      href="/docs/checks"
      eyebrow="Reference"
      title="Checks"
      lead="Shipcheck runs ten small checks, in this order. Each one passes, warns, fails or is skipped, and failures are what block shipping."
    >
      <nav aria-label="Checks on this page" className="not-prose grid grid-cols-2 gap-2 sm:grid-cols-4">
        {checkGroups.map((g) => (
          <a
            key={g.slug}
            href={`#${g.slug}`}
            className="no-prose rounded-[4px] border border-ink bg-card px-3 py-2 font-mono text-[13px] transition-[transform,box-shadow] duration-150 ease-out hover:-translate-x-px hover:-translate-y-px hover:shadow-hard-sm"
          >
            {g.label}
          </a>
        ))}
      </nav>
      <p>
        Every check can be turned off or have its <Link href="/docs/configuration#severity">severity</Link> changed
        in <code>.shipcheck.yml</code>. By default, only the problems marked <code>✗</code> below block shipping.
      </p>

      {checkGroups.map((group) =>
        group.checks.length === 1 ? (
          <section key={group.slug} aria-labelledby={group.slug}>
            <H2 id={group.slug}>{group.title}</H2>
            <div className="mt-4 space-y-4">
              <CheckBody check={group.checks[0]} />
            </div>
          </section>
        ) : (
          <section key={group.slug} aria-labelledby={group.slug}>
            <H2 id={group.slug}>{group.title}</H2>
            <p className="mt-3">{group.summary}</p>
            {group.checks.map((check) => (
              <div key={check.id} className="mt-2 space-y-4">
                <H3 id={check.id.replace(".", "-")}>{check.name}</H3>
                <CheckBody check={check} />
              </div>
            ))}
          </section>
        ),
      )}

      <H2 id="adding-checks">Adding a check</H2>
      <p>
        Checks live in <code>cli/internal/checks</code>. Each one is a small value with an ID, a name, a description
        and a <code>Run</code> function; the runner handles timing, severity and output. See{" "}
        <Link href="/contributing">Contributing</Link> for a walkthrough.
      </p>
    </DocPage>
  );
}
