package checks

import (
	"regexp"
	"strings"
)

// KeyValue is one assignment from a dotenv file.
type KeyValue struct {
	Key   string
	Value string
}

var dotenvKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)

// ParseDotenv reads KEY=value lines. It supports comments, blank lines,
// an optional `export` prefix and single- or double-quoted values.
// Malformed lines are ignored rather than guessed at.
func ParseDotenv(content string) []KeyValue {
	var out []KeyValue
	for _, raw := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if !dotenvKey.MatchString(key) {
			continue
		}
		out = append(out, KeyValue{Key: key, Value: unquote(strings.TrimSpace(value))})
	}
	return out
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') {
		if end := strings.IndexByte(v[1:], v[0]); end >= 0 {
			return v[1 : end+1]
		}
	}
	// Strip inline comments from unquoted values: KEY=value # comment
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return v
}
