import type { CSSProperties, ReactNode } from "react";
import type { Line, Tone } from "@/lib/terminal";

const symbolColor: Partial<Record<Tone, string>> = {
  pass: "text-term-pass",
  warn: "text-term-warn",
  fail: "text-term-fail",
};

function renderLine(line: Line): ReactNode {
  const { tone, text } = line;
  switch (tone) {
    case "cmd":
      return (
        <>
          <span className="text-term-dim select-none">$ </span>
          <span className="text-term-text">{text.slice(2)}</span>
        </>
      );
    case "title":
      return <span className="font-semibold tracking-wide text-white">{text}</span>;
    case "pass":
    case "warn":
    case "fail":
      return (
        <>
          <span className={symbolColor[tone]}>{text.slice(0, 1)}</span>
          {text.slice(1)}
        </>
      );
    case "ready":
      return <span className="font-semibold text-term-pass">{text}</span>;
    case "notready":
      return <span className="font-semibold text-term-fail">{text}</span>;
    case "skip":
    case "hint":
    case "rule":
      return <span className="text-term-dim">{text}</span>;
    default:
      return text;
  }
}

type TerminalProps = {
  lines: Line[];
  /** Label shown in the title bar. */
  title?: string;
  /** Accessible name for the output region. */
  label: string;
  /** Reveal lines one by one (disabled under reduced motion). */
  animate?: boolean;
  /** Show a blinking caret after the last line. */
  caret?: boolean;
  className?: string;
  children?: ReactNode;
};

export function Terminal({
  lines,
  title = "~/my-app",
  label,
  animate = false,
  caret = false,
  className = "",
  children,
}: TerminalProps) {
  return (
    <figure
      className={`overflow-hidden rounded-md border border-ink bg-term text-term-text shadow-hard-lg ${className}`}
    >
      <div className="flex items-center gap-3 border-b border-term-line px-4 py-2.5">
        <span className="flex gap-1.5" aria-hidden>
          <span className="size-2.5 rounded-full border border-term-dim/60" />
          <span className="size-2.5 rounded-full border border-term-dim/60" />
          <span className="size-2.5 rounded-full border border-term-dim/60" />
        </span>
        <figcaption className="truncate font-mono text-xs text-term-dim">{title}</figcaption>
        {children && <div className="ml-auto">{children}</div>}
      </div>
      <pre
        tabIndex={0}
        aria-label={label}
        className={`overflow-x-auto px-4 py-4 font-mono text-[12.5px] leading-[1.6] sm:px-5 sm:py-5 sm:text-[13.5px] focus-visible:outline-accent focus-visible:-outline-offset-2 ${animate ? "term-animate" : ""}`}
      >
        <code>
          {lines.map((line, i) => (
            <span
              key={i}
              className="term-line"
              data-tone={line.tone}
              style={{ "--i": i } as CSSProperties}
            >
              {line.text === "" ? " " : renderLine(line)}
              {caret && i === lines.length - 1 && <span className="term-caret" aria-hidden />}
            </span>
          ))}
        </code>
      </pre>
    </figure>
  );
}
