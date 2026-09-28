import type { ReactNode } from "react";
import { CopyButton } from "./copy-button";

export type Lang = "bash" | "yaml" | "json" | "text" | "go";

type CodeBlockProps = {
  code: string;
  lang?: Lang;
  /** Optional file name or caption shown in the header. */
  title?: string;
  className?: string;
};

/**
 * A light, bordered code block for things you type or write. Terminal
 * output uses <Terminal> instead, so the two never look alike.
 */
export function CodeBlock({ code, lang = "bash", title, className = "" }: CodeBlockProps) {
  const source = code.replace(/^\n/, "").replace(/\s+$/, "");
  return (
    <div className={`overflow-hidden rounded-md border border-ink bg-card shadow-hard-sm ${className}`}>
      <div className="flex items-center justify-between gap-3 border-b border-subtle bg-wash py-1.5 pr-1.5 pl-4">
        <span className="truncate font-mono text-xs text-muted">{title ?? lang}</span>
        <CopyButton
          text={copyText(source, lang)}
          label={title ? `Copy ${title}` : "Copy code"}
          className="border-transparent text-muted hover:border-ink hover:bg-card hover:text-ink"
        />
      </div>
      <pre
        tabIndex={0}
        className="overflow-x-auto px-4 py-3.5 font-mono text-[13px] leading-[1.65] focus-visible:-outline-offset-2"
      >
        <code>{highlight(source, lang)}</code>
      </pre>
    </div>
  );
}

/** Commands are copied without their "$ " prompts or output lines. */
function copyText(source: string, lang: Lang): string {
  if (lang !== "bash") return source;
  const lines = source.split("\n");
  if (!lines.some((l) => l.startsWith("$ "))) return source;
  return lines
    .filter((l) => l.startsWith("$ "))
    .map((l) => l.slice(2))
    .join("\n");
}

/**
 * A deliberately tiny highlighter: comments and prompts are muted, keys
 * are emphasised. Enough structure to scan, without a dependency.
 */
function highlight(source: string, lang: Lang): ReactNode[] {
  return source.split("\n").map((line, i) => (
    <span key={i} className="block min-h-[1.65em]">
      {highlightLine(line, lang)}
    </span>
  ));
}

function highlightLine(line: string, lang: Lang): ReactNode {
  const muted = (t: string) => <span className="text-muted">{t}</span>;

  if (lang === "bash" || lang === "text") {
    if (/^\s*#/.test(line)) return muted(line);
    if (line.startsWith("$ ")) {
      return (
        <>
          <span className="text-muted select-none">$ </span>
          {line.slice(2)}
        </>
      );
    }
    return line;
  }

  if (lang === "yaml") {
    if (/^\s*#/.test(line)) return muted(line);
    const m = line.match(/^(\s*(?:-\s+)?)([\w.-]+)(:)(.*)$/);
    const commentAt = (s: string) => {
      const idx = s.search(/\s#/);
      return idx >= 0 ? (
        <>
          {s.slice(0, idx)}
          {muted(s.slice(idx))}
        </>
      ) : (
        s
      );
    };
    if (m) {
      return (
        <>
          {m[1]}
          <span className="font-semibold">{m[2]}</span>
          {m[3]}
          {commentAt(m[4])}
        </>
      );
    }
    return commentAt(line);
  }

  if (lang === "json") {
    const m = line.match(/^(\s*)("[^"]+")(:)(.*)$/);
    if (m) {
      return (
        <>
          {m[1]}
          <span className="font-semibold">{m[2]}</span>
          {m[3]}
          {m[4]}
        </>
      );
    }
    return line;
  }

  if (lang === "go") {
    if (/^\s*\/\//.test(line)) return muted(line);
    return line;
  }
  return line;
}
