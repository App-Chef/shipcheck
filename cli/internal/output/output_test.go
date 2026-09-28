package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/runner"
)

func sampleReport() *runner.Report {
	return &runner.Report{
		Results: []runner.Result{
			{ID: "git.repo", Name: "Git repository", Status: checks.Pass, Message: "Git repository", Duration: 3 * time.Millisecond},
			{ID: "tests", Name: "Tests", Status: checks.Fail, Message: "Tests failed (npm test)", Suggestion: "Run `npm test` to see the full output.", Details: []string{"$ npm test  (exit 1, 1.2s)", "FAIL src/app.test.ts"}},
			{ID: "build", Name: "Production build", Status: checks.Skip, Message: "No build command detected", Suggestion: "Set commands.build in .shipcheck.yml if the project has one."},
			{ID: "ci", Name: "CI", Status: checks.Warn, Message: "No CI configuration", Suggestion: "Add CI."},
		},
		Summary:  runner.Summary{Passed: 1, Failed: 1, Skipped: 1, Warnings: 1},
		Duration: 1500 * time.Millisecond,
	}
}

func render(rep *runner.Report, st Style, verbose bool) string {
	var buf bytes.Buffer
	txt := &Text{W: &buf, Style: st, Verbose: verbose}
	txt.Header("/work/app", "")
	for _, r := range rep.Results {
		txt.Start(checks.Check{ID: r.ID, Name: r.Name})
		txt.Finish(r)
	}
	txt.Summary(rep)
	return buf.String()
}

func TestTextOutput(t *testing.T) {
	got := render(sampleReport(), Style{}, false)
	want := `
SHIPCHECK

✓ Git repository
✗ Tests failed (npm test)
  → Run ` + "`npm test`" + ` to see the full output.
– No build command detected
⚠ No CI configuration
  → Add CI.

────────────────────────────

1 passed · 1 warning · 1 failed · 1 skipped

✗ Not ready to ship

`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(got, "\x1b") {
		t.Error("plain style must not emit ANSI escapes")
	}
}

func TestTextReady(t *testing.T) {
	rep := &runner.Report{
		Results: []runner.Result{{ID: "a", Status: checks.Pass, Message: "A"}, {ID: "b", Status: checks.Warn, Message: "B"}},
		Summary: runner.Summary{Passed: 1, Warnings: 1},
	}
	got := render(rep, Style{}, false)
	if !strings.Contains(got, "1 passed · 1 warning\n") || !strings.Contains(got, "✓ Ready to ship") {
		t.Errorf("got:\n%s", got)
	}
}

func TestTextVerbose(t *testing.T) {
	got := render(sampleReport(), Style{}, true)
	for _, want := range []string{
		"project  /work/app",
		"config   none (using defaults)",
		"    FAIL src/app.test.ts",
		"  → Set commands.build", // skip suggestions only in verbose
		"3ms",
		"· 1.5s",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("verbose output missing %q:\n%s", want, got)
		}
	}
}

func TestTextColorAndLiveProgress(t *testing.T) {
	got := render(sampleReport(), Style{Color: true, Live: true}, false)
	if !strings.Contains(got, green+"✓"+reset) || !strings.Contains(got, red+"✗"+reset) {
		t.Errorf("expected colored symbols:\n%q", got)
	}
	if !strings.Contains(got, "· Git repository…") || !strings.Contains(got, "\r\x1b[K") {
		t.Errorf("expected live progress line that is cleared:\n%q", got)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, sampleReport(), "v1.0.0", false); err != nil {
		t.Fatal(err)
	}
	var got JSONReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if got.Ready || got.Version != "v1.0.0" || got.DurationMS != 1500 {
		t.Errorf("report = %+v", got)
	}
	if got.Summary != (JSONSummary{Passed: 1, Warnings: 1, Failed: 1, Skipped: 1}) {
		t.Errorf("summary = %+v", got.Summary)
	}
	if len(got.Checks) != 4 || got.Checks[1].ID != "tests" || got.Checks[1].Status != "fail" {
		t.Errorf("checks = %+v", got.Checks)
	}
	if got.Checks[1].Details != nil {
		t.Error("details must be omitted without --verbose")
	}
	if strings.Contains(buf.String(), `"suggestion": ""`) {
		t.Error("empty suggestions should be omitted")
	}

	buf.Reset()
	_ = WriteJSON(&buf, sampleReport(), "v1.0.0", true)
	_ = json.Unmarshal(buf.Bytes(), &got)
	if len(got.Checks[1].Details) != 2 {
		t.Errorf("verbose details = %q", got.Checks[1].Details)
	}
}

func TestJSONEmptyReportHasArray(t *testing.T) {
	var buf bytes.Buffer
	_ = WriteJSON(&buf, &runner.Report{}, "dev", false)
	if !strings.Contains(buf.String(), `"checks": []`) || !strings.Contains(buf.String(), `"ready": true`) {
		t.Errorf("got %s", buf.String())
	}
}

func TestJSONError(t *testing.T) {
	var buf bytes.Buffer
	WriteJSONError(&buf, errString("bad config"))
	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil || got["error"] != "bad config" {
		t.Errorf("got %s (%v)", buf.String(), err)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestDetectStyleRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("FORCE_COLOR", "1")
	if s := DetectStyle(nil, false); s.Color {
		t.Error("NO_COLOR must win")
	}
	t.Setenv("NO_COLOR", "")
	if s := DetectStyle(nil, true); s.Color {
		t.Error("--no-color must win")
	}
	if s := DetectStyle(nil, false); !s.Color || s.Live {
		t.Errorf("FORCE_COLOR should enable color without live updates on a non-terminal: %+v", s)
	}
	t.Setenv("FORCE_COLOR", "")
	if s := DetectStyle(nil, false); s.Color {
		t.Error("non-terminals should not get color by default")
	}
}

func TestStripANSIAndPad(t *testing.T) {
	st := Style{Color: true}
	s := st.Green("✓") + " ok"
	if stripANSI(s) != "✓ ok" {
		t.Errorf("stripANSI = %q", stripANSI(s))
	}
	if got := padRight(s, 6); stripANSI(got) != "✓ ok  " {
		t.Errorf("padRight = %q", stripANSI(got))
	}
}
