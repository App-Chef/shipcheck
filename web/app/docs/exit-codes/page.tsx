import type { Metadata } from "next";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2 } from "@/components/docs/doc-page";

export const metadata: Metadata = {
  title: "Exit codes",
  description: "Shipcheck exits 0 when ready, 1 when a check fails, and 2 when Shipcheck itself fails.",
};

const codes = [
  {
    code: "0",
    title: "Ready",
    body: "No check failed. Warnings and skipped checks are allowed.",
  },
  {
    code: "1",
    title: "Not ready",
    body: "At least one check failed. The report says which, and why.",
  },
  {
    code: "2",
    title: "Shipcheck error",
    body: "Shipcheck couldn't run: an invalid config file, an unknown flag, a missing directory, or an interrupt.",
  },
];

export default function ExitCodes() {
  return (
    <DocPage
      href="/docs/exit-codes"
      eyebrow="Reference"
      title="Exit codes"
      lead="Three exit codes, so scripts can tell “not ready” apart from “couldn't check”."
    >
      <div className="grid gap-3 sm:grid-cols-3">
        {codes.map((c) => (
          <div key={c.code} className="card p-4">
            <p className="font-mono text-3xl font-semibold">{c.code}</p>
            <p className="mt-2 font-semibold">{c.title}</p>
            <p className="mt-1 text-[15px] leading-relaxed text-muted">{c.body}</p>
          </div>
        ))}
      </div>

      <H2 id="gating">Gating a deploy</H2>
      <CodeBlock lang="bash" code={`$ shipcheck && echo "Ready to deploy"`} />
      <p>The deploy only runs on exit code 0.</p>

      <H2 id="distinguishing">Telling failures apart</H2>
      <CodeBlock
        lang="bash"
        title="deploy.sh"
        code={`shipcheck
case $? in
  0) ./deploy.sh ;;
  1) echo "Not ready to ship, fix the failures above." ; exit 1 ;;
  *) echo "Shipcheck could not run." ; exit 2 ;;
esac`}
      />

      <H2 id="severity">Warnings and severity</H2>
      <p>
        Warnings never change the exit code. To make a warning block shipping, set its severity to{" "}
        <code>error</code> in <code>.shipcheck.yml</code>; to stop a failure from blocking, set it to{" "}
        <code>warn</code>.
      </p>
      <CodeBlock
        lang="yaml"
        code={`severity:
  ci: error        # missing CI now exits 1
  git_clean: warn  # uncommitted changes no longer exit 1`}
      />
    </DocPage>
  );
}
