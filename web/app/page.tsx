import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { CopyButton } from "@/components/copy-button";
import { ArrowRight, GitHub } from "@/components/icons";
import { TerminalDemo } from "@/components/terminal-demo";
import { checkGroups } from "@/lib/checks";
import { site } from "@/lib/site";

const steps = [
  { n: "01", title: "Run", body: <code className="kbd-inline">shipcheck</code> },
  { n: "02", title: "Check", body: "Shipcheck inspects your project: Git, tests, build, environment and more." },
  { n: "03", title: "Ship", body: "Fix what matters and deploy with confidence." },
];

const forgotten = [
  "a file you never committed",
  "an API key missing in production",
  "a build that only breaks in production mode",
  "a version number you already released",
];

const ciSnippet = `- name: Install Shipcheck
  run: ${site.installCommand}

- name: Run Shipcheck
  run: shipcheck --json > shipcheck.json`;

function SectionHeading({ eyebrow, title, children }: { eyebrow: string; title: string; children?: React.ReactNode }) {
  return (
    <div className="max-w-2xl">
      <p className="eyebrow">{eyebrow}</p>
      <h2 className="mt-3 text-3xl leading-[1.1] font-semibold tracking-[-0.03em] text-balance sm:text-4xl">{title}</h2>
      {children && <p className="mt-4 text-lg leading-relaxed text-muted text-pretty">{children}</p>}
    </div>
  );
}

export default function Home() {
  return (
    <>
      {/* Hero */}
      <section className="border-b border-ink">
        <div className="container-page pt-14 pb-16 sm:pt-20 sm:pb-24">
          <div className="max-w-3xl">
            <p className="inline-flex items-center gap-2 rounded-full border border-ink bg-card px-3 py-1 font-mono text-xs">
              <span className="size-1.5 rounded-full bg-[#1f8a43]" aria-hidden />
              Open source · local · offline
            </p>
            <h1 className="mt-6 text-[2.75rem] leading-[1.02] font-semibold tracking-[-0.045em] sm:text-6xl lg:text-7xl">
              Know you&apos;re ready
              <br />
              before you <span className="bg-accent px-1.5 box-decoration-clone">ship.</span>
            </h1>
            <p className="mt-6 max-w-xl text-lg leading-relaxed text-muted sm:text-xl text-pretty">
              A minimal CLI that checks your project for common release blockers.
            </p>
            <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:items-center">
              <Link href="/docs/installation" className="btn btn-primary">
                Install Shipcheck
                <ArrowRight />
              </Link>
              <a href={site.repo} className="btn btn-secondary">
                <GitHub />
                View on GitHub
              </a>
            </div>
            <div className="mt-6 flex max-w-full items-center gap-2 overflow-hidden rounded-md border border-subtle bg-card py-1 pr-1 pl-3 font-mono text-[12.5px] sm:inline-flex">
              <span className="text-muted select-none" aria-hidden>$</span>
              <code className="min-w-0 flex-1 truncate">{site.installCommand}</code>
              <CopyButton
                text={site.installCommand}
                label="Copy install command"
                className="shrink-0 border-transparent text-muted hover:border-ink hover:text-ink"
              />
            </div>
          </div>

          <div className="mt-14 max-w-3xl sm:mt-16">
            <TerminalDemo />
          </div>
        </div>
      </section>

      {/* Problem */}
      <section aria-labelledby="problem" className="border-b border-ink bg-card">
        <div className="container-page grid gap-12 py-20 sm:py-24 lg:grid-cols-2 lg:gap-16">
          <div>
            <p className="eyebrow">The problem</p>
            <h2 id="problem" className="mt-5 space-y-2 text-3xl leading-[1.15] font-semibold tracking-[-0.03em] sm:text-4xl">
              <span className="block">You finished the feature.</span>
              <span className="block">Tests pass.</span>
              <span className="block">You deploy.</span>
              <span className="block text-muted">
                Then something breaks because you forgot one small thing.
              </span>
            </h2>
          </div>
          <div className="self-end">
            <p className="text-lg leading-relaxed text-muted">It&apos;s rarely the hard part that breaks a release. It&apos;s</p>
            <ul className="mt-5 divide-y divide-subtle border-y border-subtle">
              {forgotten.map((f) => (
                <li key={f} className="flex items-center gap-3 py-3.5 text-lg">
                  <span className="font-mono text-term-fail" aria-hidden>✗</span>
                  {f}
                </li>
              ))}
            </ul>
            <p className="mt-6 text-lg leading-relaxed text-muted">
              Solo developers and small teams don&apos;t have a release manager to catch these.
            </p>
          </div>
        </div>
      </section>

      {/* Solution + how it works */}
      <section aria-labelledby="how" className="border-b border-ink">
        <div className="container-page py-20 sm:py-24">
          <p className="eyebrow">The solution</p>
          <h2 id="how" className="mt-3 text-4xl leading-[1.05] font-semibold tracking-[-0.04em] sm:text-5xl">
            Run <span className="font-mono tracking-[-0.06em]">shipcheck</span> before you ship.
          </h2>
          <ol className="mt-12 grid gap-5 md:grid-cols-3">
            {steps.map((s) => (
              <li key={s.n} className="card p-6">
                <p className="font-mono text-sm text-muted">{s.n}</p>
                <h3 className="mt-6 text-2xl font-semibold tracking-tight">{s.title}</h3>
                <p className="mt-2 leading-relaxed text-muted">{s.body}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* Checks */}
      <section aria-labelledby="checks" className="border-b border-ink bg-wash">
        <div className="container-page py-20 sm:py-24">
          <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
            <SectionHeading eyebrow="Checks" title="Everything that is easy to forget.">
              Test and build commands are detected for Go, Node.js, Python, Java, PHP and Rust. No setup required.
            </SectionHeading>
            <Link href="/docs/checks" className="link flex shrink-0 items-center gap-1.5 font-medium">
              All checks <ArrowRight />
            </Link>
          </div>
          <ul className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {checkGroups.map((g) => (
              <li key={g.slug}>
                <Link
                  href={`/docs/checks#${g.slug}`}
                  className="card card-interactive flex h-full flex-col p-5"
                >
                  <h3 className="text-lg font-semibold tracking-tight">{g.label}</h3>
                  <p className="mt-2 flex-1 text-[15px] leading-relaxed text-muted">{g.summary}</p>
                  <p className="mt-5 truncate rounded-[4px] bg-term px-2.5 py-1.5 font-mono text-xs text-term-text">
                    <span className={g.example.startsWith("✓") ? "text-term-pass" : "text-term-warn"}>
                      {g.example.slice(0, 1)}
                    </span>
                    {g.example.slice(1)}
                  </p>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* CI */}
      <section aria-labelledby="ci" className="border-b border-ink">
        <div className="container-page grid items-center gap-12 py-20 sm:py-24 lg:grid-cols-[1fr_1.1fr] lg:gap-16">
          <div>
            <p className="eyebrow">CI</p>
            <h2 id="ci" className="mt-3 text-3xl leading-[1.1] font-semibold tracking-[-0.03em] sm:text-4xl">
              The same checks, on every push.
            </h2>
            <p className="mt-4 text-lg leading-relaxed text-muted">
              Shipcheck exits <code className="kbd-inline">0</code> when you&apos;re ready and{" "}
              <code className="kbd-inline">1</code> when you&apos;re not, so any CI system can run it.{" "}
              <code className="kbd-inline">--json</code> prints a clean report for scripts and build artifacts.
            </p>
            <ul className="mt-6 space-y-2 text-[15px]">
              <li className="flex gap-2.5"><span className="font-mono text-[#1f8a43]" aria-hidden>✓</span>Stable exit codes: 0 ready, 1 not ready, 2 error</li>
              <li className="flex gap-2.5"><span className="font-mono text-[#1f8a43]" aria-hidden>✓</span>JSON only on stdout, never mixed with text</li>
              <li className="flex gap-2.5"><span className="font-mono text-[#1f8a43]" aria-hidden>✓</span>Secrets are checked for presence, never printed</li>
            </ul>
            <Link href="/docs/ci" className="link mt-8 inline-flex items-center gap-1.5 font-medium">
              CI guide <ArrowRight />
            </Link>
          </div>
          <div className="min-w-0 space-y-4">
            <CodeBlock lang="yaml" title=".github/workflows/ci.yml" code={ciSnippet} />
            <CodeBlock
              lang="bash"
              title="terminal"
              code={`$ shipcheck --json | jq '.ready'
false
$ shipcheck && ./deploy.sh`}
            />
          </div>
        </div>
      </section>

      {/* Open source */}
      <section aria-labelledby="open-source" className="border-b border-ink bg-card">
        <div className="container-page py-20 sm:py-24">
          <SectionHeading eyebrow="Open source" title="Yours to read, run and change.">
            Shipcheck is MIT licensed and runs entirely on your machine. There&apos;s nothing to sign up for.
          </SectionHeading>
          <dl className="mt-12 grid gap-px overflow-hidden rounded-md border border-ink bg-ink sm:grid-cols-2 lg:grid-cols-4">
            {[
              ["No account", "No sign-up, no API key, no login."],
              ["Offline", "No network access. Works on a plane."],
              ["No telemetry", "Nothing about you or your code is collected."],
              ["Read-only", "Never edits, deletes or commits your files."],
            ].map(([t, d]) => (
              <div key={t} className="bg-card p-5">
                <dt className="font-semibold">{t}</dt>
                <dd className="mt-1.5 text-[15px] leading-relaxed text-muted">{d}</dd>
              </div>
            ))}
          </dl>
          <div className="mt-8 flex flex-wrap gap-x-6 gap-y-3 font-medium">
            <a href={site.repo} className="link inline-flex items-center gap-2">
              <GitHub /> Source on GitHub
            </a>
            <Link href="/contributing" className="link">Contributing guide</Link>
            <Link href="/docs/security" className="link">Security model</Link>
          </div>
        </div>
      </section>

      {/* Final CTA */}
      <section aria-labelledby="cta">
        <div className="container-page py-20 sm:py-24">
          <div className="card flex flex-col items-start gap-6 bg-accent p-8 shadow-hard-lg sm:p-12 md:flex-row md:items-center md:justify-between">
            <div>
              <h2 id="cta" className="text-4xl leading-[1.05] font-semibold tracking-[-0.04em] sm:text-5xl">
                Ready to ship?
              </h2>
              <p className="mt-3 text-lg">Install Shipcheck and run one command.</p>
            </div>
            <Link href="/docs/quick-start" className="btn bg-ink text-paper">
              Get started
              <ArrowRight />
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}
