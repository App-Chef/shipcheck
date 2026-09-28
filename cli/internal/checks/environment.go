package checks

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
)

// EnvTemplates are the files that document required variables.
var EnvTemplates = []string{".env.example", ".env.sample", ".example.env"}

// Environment checks that required environment variables are set.
// Values are never printed, logged or stored.
var Environment = Check{
	ID:          "environment",
	Name:        "Environment",
	Description: "Checks that required environment variables are set, without exposing their values.",
	Run: func(_ context.Context, env *Env) Result {
		template := env.EnvTemplate()
		required, err := env.RequiredEnv()
		if err != nil {
			return Result{Status: Fail, Message: "Could not read " + template, Details: []string{err.Error()}}
		}

		values := env.envValues()
		var missing []string
		for _, name := range required {
			if strings.TrimSpace(values[name]) == "" {
				missing = append(missing, name)
			}
		}
		// A committed .env file is a leaked secret waiting to happen.
		tracked := env.trackedEnvFiles()
		var trackedNote []string
		if len(tracked) > 0 {
			trackedNote = []string{strings.Join(tracked, ", ") + " is committed to Git"}
		}

		switch {
		case len(missing) > 0:
			return Result{
				Status:     Fail,
				Message:    plural(len(missing), "environment variable missing", "environment variables missing"),
				Suggestion: fmt.Sprintf("Set %s in your environment or in %s.", strings.Join(missing, ", "), strings.Join(env.Config.EnvFiles(), ", ")),
				Details:    append(prefixAll("missing: ", missing), trackedNote...),
			}
		case len(tracked) > 0:
			return Result{
				Status:     Warn,
				Message:    trackedNote[0],
				Suggestion: "Remove it with `git rm --cached` and add it to .gitignore. Rotate any secrets it contained.",
			}
		case len(required) == 0 && template == "":
			return skip("No environment variables required")
		case len(required) == 0:
			return pass("Environment checked (" + template + " lists no variables)")
		}
		return pass("Environment checked", plural(len(required), "variable", "variables")+" set")
	},
}

// EnvTemplate returns the first env template file present, or "".
func (e *Env) EnvTemplate() string {
	for _, name := range EnvTemplates {
		if e.Project.IsFile(name) {
			return name
		}
	}
	return ""
}

// RequiredEnv returns the variable names from the env template plus
// environment.required, de-duplicated, in a stable order.
func (e *Env) RequiredEnv() ([]string, error) {
	seen := map[string]bool{}
	var names []string
	add := func(n string) {
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	if tpl := e.EnvTemplate(); tpl != "" {
		data, err := e.Project.ReadFile(tpl)
		if err != nil {
			return nil, err
		}
		for _, kv := range ParseDotenv(string(data)) {
			add(kv.Key)
		}
	}
	for _, n := range e.Config.Environment.Required {
		add(n)
	}
	return names, nil
}

// envValues merges values from the configured dotenv files with the
// process environment. The process environment wins.
func (e *Env) envValues() map[string]string {
	values := map[string]string{}
	for _, f := range e.Config.EnvFiles() {
		data, err := e.Project.ReadFile(f)
		if err != nil {
			continue
		}
		for _, kv := range ParseDotenv(string(data)) {
			values[kv.Key] = kv.Value
		}
	}
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok && v != "" {
			values[k] = v
		}
	}
	return values
}

// trackedEnvFiles returns dotenv files in the project root that are
// committed to Git. Templates are expected to be committed and ignored.
func (e *Env) trackedEnvFiles() []string {
	git := e.Project.Git
	if !git.Available() || !git.IsRepo() {
		return nil
	}
	entries, err := os.ReadDir(e.Project.Root)
	if err != nil {
		return nil
	}
	var tracked []string
	for _, entry := range entries {
		name := entry.Name()
		if !entry.Type().IsRegular() || !isDotenvName(name) {
			continue
		}
		if git.IsTracked(name) {
			tracked = append(tracked, name)
		}
	}
	sort.Strings(tracked)
	return tracked
}

func isDotenvName(name string) bool {
	if name != ".env" && !strings.HasPrefix(name, ".env.") {
		return false
	}
	lower := strings.ToLower(name)
	for _, marker := range []string{"example", "sample", "template", "dist", "defaults", "schema"} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return true
}

// Redact replaces the values of required variables in s with "****".
// It is applied to any command output Shipcheck displays.
func (e *Env) Redact(s string) string {
	e.secretsOnce.Do(func() {
		names, err := e.RequiredEnv()
		if err != nil {
			return
		}
		values := e.envValues()
		for _, n := range names {
			// Very short values would redact ordinary words.
			if v := values[n]; len(v) >= 4 {
				e.secrets = append(e.secrets, v)
			}
		}
		// Replace longer values first so overlapping secrets stay hidden.
		sort.Slice(e.secrets, func(i, j int) bool { return len(e.secrets[i]) > len(e.secrets[j]) })
	})
	for _, v := range e.secrets {
		s = strings.ReplaceAll(s, v, "****")
	}
	return s
}

func prefixAll(prefix string, items []string) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = prefix + it
	}
	return out
}
