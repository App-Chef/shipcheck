import Link from "next/link";
import { LogoMark } from "./logo";
import { repoFile, site } from "@/lib/site";

const columns = [
  {
    title: "Docs",
    links: [
      { href: "/docs/installation", label: "Installation" },
      { href: "/docs/quick-start", label: "Quick start" },
      { href: "/docs/configuration", label: "Configuration" },
      { href: "/docs/checks", label: "Checks" },
    ],
  },
  {
    title: "Reference",
    links: [
      { href: "/docs/ci", label: "CI" },
      { href: "/docs/json-output", label: "JSON output" },
      { href: "/docs/exit-codes", label: "Exit codes" },
      { href: "/docs/security", label: "Security model" },
    ],
  },
  {
    title: "Project",
    links: [
      { href: site.repo, label: "GitHub" },
      { href: "/contributing", label: "Contributing" },
      { href: repoFile("CODE_OF_CONDUCT.md"), label: "Code of conduct" },
      { href: repoFile("LICENSE"), label: "MIT License" },
    ],
  },
];

export function SiteFooter() {
  return (
    <footer className="border-t border-ink bg-paper">
      <div className="container-page grid gap-10 py-12 sm:grid-cols-2 lg:grid-cols-[1.4fr_1fr_1fr_1fr]">
        <div className="max-w-xs">
          <Link href="/" className="flex items-center gap-2.5 font-semibold tracking-tight">
            <LogoMark />
            Shipcheck
          </Link>
          <p className="mt-4 text-sm leading-relaxed text-muted">
            Know you&apos;re ready before you ship. Open source, local and offline. No account,
            no telemetry.
          </p>
        </div>
        {columns.map((col) => (
          <nav key={col.title} aria-label={col.title}>
            <h2 className="eyebrow">{col.title}</h2>
            <ul className="mt-4 space-y-2.5 text-sm">
              {col.links.map((l) => (
                <li key={l.href}>
                  {l.href.startsWith("http") ? (
                    <a href={l.href} className="text-muted transition-colors duration-150 hover:text-ink">
                      {l.label}
                    </a>
                  ) : (
                    <Link href={l.href} className="text-muted transition-colors duration-150 hover:text-ink">
                      {l.label}
                    </Link>
                  )}
                </li>
              ))}
            </ul>
          </nav>
        ))}
      </div>
      <div className="border-t border-subtle">
        <div className="container-page flex flex-col gap-2 py-5 font-mono text-xs text-muted sm:flex-row sm:justify-between">
          <p>Released under the {site.license} License.</p>
          <p>Built for solo developers and small teams.</p>
        </div>
      </div>
    </footer>
  );
}
