import Link from "next/link";
import type { ReactNode } from "react";
import { ArrowLeft, ArrowRight, Hash } from "../icons";
import { docNeighbours } from "@/lib/docs";
import { repoFile } from "@/lib/site";

type DocPageProps = {
  href: string;
  eyebrow?: string;
  title: string;
  lead: ReactNode;
  children: ReactNode;
};

export function DocPage({ href, eyebrow = "Documentation", title, lead, children }: DocPageProps) {
  const { prev, next } = docNeighbours(href);
  return (
    <article className="min-w-0">
      <header className="border-b border-subtle pb-8">
        <p className="eyebrow">{eyebrow}</p>
        <h1 className="mt-3 text-[2rem] leading-[1.1] font-semibold tracking-[-0.03em] text-balance sm:text-[2.5rem]">
          {title}
        </h1>
        <p className="mt-4 max-w-2xl text-lg leading-relaxed text-muted text-pretty">{lead}</p>
      </header>

      <div className="prose-doc mt-8 max-w-[46rem]">{children}</div>

      <nav aria-label="Previous and next pages" className="mt-16 grid gap-4 border-t border-subtle pt-8 sm:grid-cols-2">
        {prev ? (
          <Link href={prev.href} className="card card-interactive group flex flex-col gap-1 p-4">
            <span className="flex items-center gap-1.5 font-mono text-xs text-muted">
              <ArrowLeft width={13} height={13} /> Previous
            </span>
            <span className="font-medium">{prev.title}</span>
          </Link>
        ) : (
          <span className="hidden sm:block" />
        )}
        {next && (
          <Link href={next.href} className="card card-interactive group flex flex-col items-end gap-1 p-4 text-right">
            <span className="flex items-center gap-1.5 font-mono text-xs text-muted">
              Next <ArrowRight width={13} height={13} />
            </span>
            <span className="font-medium">{next.title}</span>
          </Link>
        )}
      </nav>
      <p className="mt-8 font-mono text-xs text-muted">
        Found a mistake?{" "}
        <a className="link" href={repoFile(`web/app${href}/page.tsx`)}>
          Edit this page on GitHub
        </a>
      </p>
    </article>
  );
}

/** A section heading with a stable anchor link. */
export function H2({ id, children }: { id: string; children: ReactNode }) {
  return (
    <h2 id={id} className="group relative">
      {children}
      <a
        href={`#${id}`}
        className="no-prose ml-2 inline-flex translate-y-[-1px] align-middle text-muted opacity-0 transition-opacity duration-150 ease-out group-hover:opacity-100 focus-visible:opacity-100"
        aria-label={`Link to section: ${typeof children === "string" ? children : id}`}
      >
        <Hash width={16} height={16} />
      </a>
    </h2>
  );
}

export function H3({ id, children }: { id: string; children: ReactNode }) {
  return (
    <h3 id={id} className="group relative">
      {children}
      <a
        href={`#${id}`}
        className="no-prose ml-2 inline-flex text-muted opacity-0 transition-opacity duration-150 ease-out group-hover:opacity-100 focus-visible:opacity-100"
        aria-label={`Link to section: ${typeof children === "string" ? children : id}`}
      >
        <Hash width={14} height={14} />
      </a>
    </h3>
  );
}

/** A quiet note for tips and caveats. */
export function Note({ title = "Note", children }: { title?: string; children: ReactNode }) {
  return (
    <aside className="rounded-md border border-ink border-l-4 bg-card px-4 py-3.5 text-[15px] leading-relaxed">
      <p className="font-mono text-xs font-medium tracking-wider text-muted uppercase">{title}</p>
      <div className="mt-1.5 [&>*+*]:mt-2">{children}</div>
    </aside>
  );
}

export function TableWrap({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-md border border-ink bg-card" tabIndex={0} role="region" aria-label="Table">
      {children}
    </div>
  );
}
