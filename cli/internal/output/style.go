// Package output renders reports for terminals and machines.
package output

import (
	"os"
	"strings"
)

// Style decides whether ANSI escapes are used.
type Style struct {
	Color bool
	// Live allows rewriting the current line to show progress.
	Live bool
}

// DetectStyle picks a style for f, honouring NO_COLOR, FORCE_COLOR,
// TERM=dumb and the --no-color flag.
func DetectStyle(f *os.File, noColor bool) Style {
	tty := isTerminal(f)
	if noColor || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return Style{}
	}
	if v := os.Getenv("FORCE_COLOR"); v != "" && v != "0" && !strings.EqualFold(v, "false") {
		enableVirtualTerminal(f)
		return Style{Color: true, Live: tty}
	}
	if !tty || !enableVirtualTerminal(f) {
		return Style{}
	}
	return Style{Color: true, Live: true}
}

func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
)

func (s Style) wrap(code, text string) string {
	if !s.Color || text == "" {
		return text
	}
	return code + text + reset
}

func (s Style) Bold(t string) string   { return s.wrap(bold, t) }
func (s Style) Dim(t string) string    { return s.wrap(dim, t) }
func (s Style) Red(t string) string    { return s.wrap(red, t) }
func (s Style) Green(t string) string  { return s.wrap(green, t) }
func (s Style) Yellow(t string) string { return s.wrap(yellow, t) }
