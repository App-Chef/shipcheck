package output

import (
	"encoding/json"
	"io"

	"github.com/App-Chef/shipcheck/cli/internal/runner"
)

// JSONReport is the stable machine-readable report printed by --json.
type JSONReport struct {
	Ready      bool        `json:"ready"`
	Version    string      `json:"version"`
	Summary    JSONSummary `json:"summary"`
	DurationMS int64       `json:"duration_ms"`
	Checks     []JSONCheck `json:"checks"`
}

// JSONSummary counts results by status.
type JSONSummary struct {
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
}

// JSONCheck is one check result.
type JSONCheck struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Message     string   `json:"message"`
	Suggestion  string   `json:"suggestion,omitempty"`
	DurationMS  int64    `json:"duration_ms"`
	Details     []string `json:"details,omitempty"`
}

// WriteJSON writes the report as indented JSON. Details (such as the tail
// of failing test output) are included only when verbose is set.
func WriteJSON(w io.Writer, rep *runner.Report, version string, verbose bool) error {
	out := JSONReport{
		Ready:   rep.Ready(),
		Version: version,
		Summary: JSONSummary{
			Passed:   rep.Summary.Passed,
			Warnings: rep.Summary.Warnings,
			Failed:   rep.Summary.Failed,
			Skipped:  rep.Summary.Skipped,
		},
		DurationMS: rep.Duration.Milliseconds(),
		Checks:     make([]JSONCheck, 0, len(rep.Results)),
	}
	for _, r := range rep.Results {
		c := JSONCheck{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Status:      string(r.Status),
			Message:     r.Message,
			Suggestion:  r.Suggestion,
			DurationMS:  r.Duration.Milliseconds(),
		}
		if verbose {
			c.Details = r.Details
		}
		out.Checks = append(out.Checks, c)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

// WriteJSONError reports a Shipcheck error (exit code 2) as JSON so that
// --json consumers never have to parse plain text.
func WriteJSONError(w io.Writer, err error) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]string{"error": err.Error()})
}
