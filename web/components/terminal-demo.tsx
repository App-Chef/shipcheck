"use client";

import { useId, useRef, useState, type KeyboardEvent } from "react";
import { Terminal } from "./terminal";
import {
  jsonTranscript,
  notReadyTranscript,
  parseTranscript,
  readyTranscript,
} from "@/lib/terminal";

const scenarios = [
  { id: "ready", label: "Ready", exit: 0, lines: parseTranscript(readyTranscript) },
  { id: "blocked", label: "Not ready", exit: 1, lines: parseTranscript(notReadyTranscript) },
  { id: "json", label: "--json", exit: 1, lines: parseTranscript(jsonTranscript) },
] as const;

/**
 * The homepage terminal. Output is real CLI output; switching tabs replays
 * the line reveal. With reduced motion, lines appear immediately.
 */
export function TerminalDemo() {
  const [active, setActive] = useState(0);
  const [run, setRun] = useState(0);
  const tabs = useRef<(HTMLButtonElement | null)[]>([]);
  const baseId = useId();
  const scenario = scenarios[active];

  function select(i: number) {
    setActive(i);
    setRun((r) => r + 1);
  }

  function onKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    const last = scenarios.length - 1;
    let next: number | null = null;
    if (e.key === "ArrowRight") next = active === last ? 0 : active + 1;
    if (e.key === "ArrowLeft") next = active === 0 ? last : active - 1;
    if (e.key === "Home") next = 0;
    if (e.key === "End") next = last;
    if (next === null) return;
    e.preventDefault();
    select(next);
    tabs.current[next]?.focus();
  }

  return (
    <div>
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div
          role="tablist"
          aria-label="Example runs"
          onKeyDown={onKeyDown}
          className="inline-flex rounded-md border border-ink bg-card p-0.5 shadow-hard-sm"
        >
          {scenarios.map((s, i) => {
            const selected = i === active;
            return (
              <button
                key={s.id}
                ref={(el) => {
                  tabs.current[i] = el;
                }}
                id={`${baseId}-tab-${s.id}`}
                role="tab"
                type="button"
                aria-selected={selected}
                aria-controls={`${baseId}-panel`}
                tabIndex={selected ? 0 : -1}
                onClick={() => select(i)}
                className={`rounded-[4px] px-3 py-1.5 font-mono text-xs font-medium transition-colors duration-150 ease-out ${
                  selected ? "bg-ink text-paper" : "text-muted hover:text-ink"
                }`}
              >
                {s.label}
              </button>
            );
          })}
        </div>
        <p className="font-mono text-xs text-muted">
          exit code <span className="text-ink">{scenario.exit}</span>
        </p>
      </div>
      <div
        id={`${baseId}-panel`}
        role="tabpanel"
        aria-labelledby={`${baseId}-tab-${scenario.id}`}
      >
        <Terminal
          key={`${scenario.id}-${run}`}
          lines={scenario.lines}
          label={`Shipcheck output: ${scenario.label}`}
          animate
        />
      </div>
    </div>
  );
}
