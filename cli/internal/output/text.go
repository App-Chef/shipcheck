package output

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/runner"
)

// Text renders a report for humans, streaming one line per check.
type Text struct {
	W       io.Writer
	Style   Style
	Verbose bool

	pending bool
}

// Header prints the title. In verbose mode it also shows what is
// being checked and which config file is in use.
func (t *Text) Header(root, configPath string) {
	fmt.Fprintln(t.W)
	fmt.Fprintln(t.W, t.Style.Bold("SHIPCHECK"))
	if t.Verbose {
		fmt.Fprintln(t.W, t.Style.Dim("project  "+root))
		if configPath == "" {
			configPath = "none (using defaults)"
		}
		fmt.Fprintln(t.W, t.Style.Dim("config   "+configPath))
	}
	fmt.Fprintln(t.W)
}

// Start shows a transient progress line on interactive terminals.
func (t *Text) Start(c checks.Check) {
	if !t.Style.Live {
		return
	}
	fmt.Fprint(t.W, t.Style.Dim("· "+c.Name+"…"))
	t.pending = true
}

// Finish replaces the progress line with the check's result.
func (t *Text) Finish(r runner.Result) {
	if t.pending {
		fmt.Fprint(t.W, "\r\x1b[K")
		t.pending = false
	}
	line := t.symbol(r.Status) + " " + t.message(r)
	if t.Verbose {
		line = padRight(line, 52) + t.Style.Dim(formatDuration(r.Duration))
	}
	fmt.Fprintln(t.W, line)

	showSuggestion := r.Suggestion != "" && (r.Status == checks.Fail || r.Status == checks.Warn || t.Verbose)
	if showSuggestion {
		fmt.Fprintln(t.W, t.Style.Dim("  → "+r.Suggestion))
	}
	if t.Verbose {
		for _, d := range r.Details {
			fmt.Fprintln(t.W, t.Style.Dim("    "+d))
		}
	}
}

// Summary prints the totals and the verdict.
func (t *Text) Summary(rep *runner.Report) {
	fmt.Fprintln(t.W)
	fmt.Fprintln(t.W, t.Style.Dim(strings.Repeat("─", 28)))
	fmt.Fprintln(t.W)

	s := rep.Summary
	parts := []string{fmt.Sprintf("%d passed", s.Passed)}
	if s.Warnings > 0 {
		parts = append(parts, pluralize(s.Warnings, "warning", "warnings"))
	}
	if s.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", s.Failed))
	}
	if s.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", s.Skipped))
	}
	summary := strings.Join(parts, " · ")
	if t.Verbose {
		summary += t.Style.Dim(" · " + formatDuration(rep.Duration))
	}
	fmt.Fprintln(t.W, summary)
	fmt.Fprintln(t.W)

	if rep.Ready() {
		fmt.Fprintln(t.W, t.Style.Bold(t.Style.Green("✓ Ready to ship")))
	} else {
		fmt.Fprintln(t.W, t.Style.Bold(t.Style.Red("✗ Not ready to ship")))
	}
	fmt.Fprintln(t.W)
}

func (t *Text) symbol(s checks.Status) string {
	switch s {
	case checks.Pass:
		return t.Style.Green("✓")
	case checks.Warn:
		return t.Style.Yellow("⚠")
	case checks.Fail:
		return t.Style.Red("✗")
	}
	return t.Style.Dim("–")
}

func (t *Text) message(r runner.Result) string {
	if r.Status == checks.Skip {
		return t.Style.Dim(r.Message)
	}
	return r.Message
}

// padRight pads s to width visible columns, ignoring ANSI escapes.
func padRight(s string, width int) string {
	visible := utf8.RuneCountInString(stripANSI(s))
	if visible >= width {
		return s + "  "
	}
	return s + strings.Repeat(" ", width-visible)
}

func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func formatDuration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return "<1ms"
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func pluralize(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
