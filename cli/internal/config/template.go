package config

import (
	"fmt"
	"strings"
)

// TemplateData describes what `shipcheck init` detected in the project.
type TemplateData struct {
	// Branch is the current git branch, used as the protected branch.
	Branch string
	// EnvTemplate is the detected .env.example-style file, if any.
	EnvTemplate string
	// TestCommands and BuildCommands are the detected commands.
	TestCommands  []string
	BuildCommands []string
}

// Template renders a commented .shipcheck.yml for a project.
func Template(d TemplateData) []byte {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }

	w("# Shipcheck configuration")
	w("# Docs: https://github.com/App-Chef/shipcheck#configuration")
	w("version: 1")
	w("")
	w("# Turn individual checks on or off. Every check is on by default.")
	w("checks:")
	w("  git_repo: true")
	w("  git_clean: true")
	w("  git_branch: true")
	w("  tests: true")
	w("  build: true")
	w("  environment: true")
	w("  readme: true")
	w("  license: true")
	w("  version: true")
	w("  ci: true")
	w("")
	w("# Change how problems are reported: \"error\" blocks shipping,")
	w("# \"warn\" is reported but does not block.")
	w("severity:")
	w("  ci: warn")
	w("  # license: error")
	w("")

	branches := []string{"main", "master"}
	if d.Branch != "" && d.Branch != "main" && d.Branch != "master" {
		branches = append([]string{d.Branch}, branches...)
	}
	w("git:")
	w("  # Warn when you are not on one of the branches you ship from.")
	w("  protected_branches:")
	for _, br := range branches {
		w("    - %s", yamlString(br))
	}
	w("")

	w("environment:")
	if d.EnvTemplate != "" {
		w("  # Variables listed in %s are checked automatically.", d.EnvTemplate)
		w("  # Add anything else that must be set before you ship:")
	} else {
		w("  # Variables that must be set before you ship. Values are never printed.")
	}
	w("  required: []")
	w("  #   - DATABASE_URL")
	w("")

	w("# Shipcheck detects test and build commands automatically.")
	w("# Set them here to override detection. Commands run without a shell.")
	w("commands:")
	writeCommands(w, "tests", d.TestCommands)
	writeCommands(w, "build", d.BuildCommands)
	w("  timeout: 10m")

	return []byte(b.String())
}

func writeCommands(w func(string, ...any), key string, detected []string) {
	switch len(detected) {
	case 0:
		w("  # %s: make %s", key, map[string]string{"tests": "test", "build": "build"}[key])
	case 1:
		w("  # %s: %s   (detected)", key, yamlString(detected[0]))
	default:
		w("  # %s:   (detected)", key)
		for _, c := range detected {
			w("  #   - %s", yamlString(c))
		}
	}
}

// yamlString quotes s when YAML would otherwise misread it.
func yamlString(s string) string {
	if s == "" || strings.ContainsAny(s, ":#{}[],&*!|>'\"%@`") || strings.TrimSpace(s) != s {
		return fmt.Sprintf("%q", s)
	}
	return s
}
