export type Tone =
  | "cmd"
  | "title"
  | "pass"
  | "warn"
  | "fail"
  | "skip"
  | "hint"
  | "rule"
  | "ready"
  | "notready"
  | "plain";

export type Line = { tone: Tone; text: string };

/**
 * Turns captured CLI output into styled lines. The transcripts below are
 * copied verbatim from real `shipcheck` runs, so the site shows exactly
 * what the CLI prints.
 */
export function parseTranscript(raw: string): Line[] {
  return raw
    .replace(/^\n/, "")
    .replace(/\n$/, "")
    .split("\n")
    .map((text): Line => {
      if (text.startsWith("$ ")) return { tone: "cmd", text };
      if (text === "SHIPCHECK") return { tone: "title", text };
      if (text.startsWith("✓ Ready to ship")) return { tone: "ready", text };
      if (text.startsWith("✗ Not ready")) return { tone: "notready", text };
      if (text.startsWith("✓")) return { tone: "pass", text };
      if (text.startsWith("⚠")) return { tone: "warn", text };
      if (text.startsWith("✗")) return { tone: "fail", text };
      if (text.startsWith("–")) return { tone: "skip", text };
      if (text.startsWith("  →") || text.startsWith("    "))
        return { tone: "hint", text };
      if (text.startsWith("───")) return { tone: "rule", text };
      return { tone: "plain", text };
    });
}

export const readyTranscript = `
$ shipcheck

SHIPCHECK

✓ Git repository
✓ Working tree clean
✓ On branch main
✓ Tests passed
✓ Production build passed
✓ Environment checked
✓ README found
✓ License found (MIT)
✓ Version 1.4.0
⚠ No CI configuration
  → Add a CI workflow so every push is tested (shipcheck runs in CI too).

────────────────────────────

9 passed · 1 warning

✓ Ready to ship
`;

export const notReadyTranscript = `
$ shipcheck

SHIPCHECK

✓ Git repository
✗ 1 uncommitted change
  → Commit or stash your changes so you ship exactly what is in Git.
✓ On branch main
✗ Tests failed (go test ./...)
  → Run \`go test ./...\` to see the full output.
✓ Production build passed
✗ 1 environment variable missing
  → Set API_KEY in your environment or in .env.
✓ README found
✓ License found (MIT)
✓ Version 1.4.0
⚠ No CI configuration
  → Add a CI workflow so every push is tested (shipcheck runs in CI too).

────────────────────────────

6 passed · 1 warning · 3 failed

✗ Not ready to ship
`;

export const jsonTranscript = `
$ shipcheck --json | jq '{ready, summary}'
{
  "ready": false,
  "summary": {
    "passed": 6,
    "warnings": 1,
    "failed": 3,
    "skipped": 0
  }
}
$ echo $?
1
`;

export const verboseTranscript = `
$ shipcheck --verbose

SHIPCHECK
project  /home/you/demo
config   none (using defaults)

✓ Git repository                                    5ms
✗ 1 uncommitted change                              11ms
  → Commit or stash your changes so you ship exactly what is in Git.
    ?? notes.txt
✓ On branch main                                    10ms
✗ Tests failed (go test ./...)                      284ms
  → Run \`go test ./...\` to see the full output.
    $ go test ./...  (exit 1, 284ms)
    --- FAIL: TestAdd (0.00s)
        demo_test.go:7: expected 5
    FAIL
    FAIL	example.com/demo	0.003s
    FAIL
✓ Production build passed                           13ms
    $ go build ./...  (13ms)
`;
