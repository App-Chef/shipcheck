package config

import (
	"errors"
	"fmt"
	"strings"
)

// shellOperators are rejected outside quotes. Shipcheck never runs a
// shell, so these would not do what the author expects; failing loudly is
// better than passing them to the program as literal arguments.
const shellOperators = ";&|<>`$(){}*?~\n\r"

// ParseCommand splits a command string into arguments without invoking a
// shell. Single and double quotes group words; shell operators such as
// pipes, redirects, variable expansion and command chaining are rejected.
func ParseCommand(s string) ([]string, error) {
	if strings.TrimSpace(s) == "" {
		return nil, errors.New("command cannot be empty")
	}
	var (
		args    []string
		current strings.Builder
		inWord  bool
		quote   rune
	)
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
			inWord = true
		case r == ' ' || r == '\t':
			if inWord {
				args = append(args, current.String())
				current.Reset()
				inWord = false
			}
		case strings.ContainsRune(shellOperators, r):
			return nil, fmt.Errorf("%q: shell syntax (%q) is not supported; commands run without a shell, so put pipelines in a script (for example an npm script or Makefile target)", s, r)
		default:
			current.WriteRune(r)
			inWord = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("%q: unterminated quote", s)
	}
	if inWord {
		args = append(args, current.String())
	}
	return args, nil
}
