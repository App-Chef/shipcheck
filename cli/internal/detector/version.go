package detector

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/App-Chef/shipcheck/cli/internal/project"
)

// Version is a project version and where it was declared.
type Version struct {
	Value  string
	Source string
}

var (
	tomlSection   = regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)
	tomlVersion   = regexp.MustCompile(`^\s*version\s*=\s*["']([^"']+)["']`)
	gradleVersion = regexp.MustCompile(`(?m)^\s*version\s*=?\s*["']([^"']+)["']`)
	pomVersion    = regexp.MustCompile(`<version>\s*([^<\s]+)\s*</version>`)
	pomParent     = regexp.MustCompile(`(?s)<parent>.*?</parent>`)
	pomDeps       = regexp.MustCompile(`(?s)<(dependencies|dependencyManagement|build|profiles|plugins)>.*?</(dependencies|dependencyManagement|build|profiles|plugins)>`)
)

// DetectVersion looks for a version declared in the project's manifest.
// It returns false when no manifest declares one.
func DetectVersion(p *project.Project) (Version, bool) {
	if pkg, ok := readPackageJSON(p); ok && pkg.Version != "" {
		return Version{pkg.Version, "package.json"}, true
	}
	if v := tomlSectionVersion(p, "Cargo.toml", "package", "workspace.package"); v != "" {
		return Version{v, "Cargo.toml"}, true
	}
	if v := tomlSectionVersion(p, "pyproject.toml", "project", "tool.poetry"); v != "" {
		return Version{v, "pyproject.toml"}, true
	}
	if data, err := p.ReadFile("composer.json"); err == nil {
		var composer struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &composer) == nil && composer.Version != "" {
			return Version{composer.Version, "composer.json"}, true
		}
	}
	if data, err := p.ReadFile("pom.xml"); err == nil {
		// Strip sections whose <version> tags belong to other artifacts.
		s := pomParent.ReplaceAllString(string(data), "")
		s = pomDeps.ReplaceAllString(s, "")
		if m := pomVersion.FindStringSubmatch(s); m != nil && !strings.Contains(m[1], "${") {
			return Version{m[1], "pom.xml"}, true
		}
	}
	for _, f := range []string{"build.gradle", "build.gradle.kts"} {
		if data, err := p.ReadFile(f); err == nil {
			if m := gradleVersion.FindSubmatch(data); m != nil {
				return Version{string(m[1]), f}, true
			}
		}
	}
	if name := p.FindFile("VERSION", "VERSION.txt"); name != "" {
		if data, err := p.ReadFile(name); err == nil {
			if v := strings.TrimSpace(string(data)); v != "" && !strings.ContainsAny(v, "\n ") {
				return Version{v, name}, true
			}
		}
	}
	return Version{}, false
}

// tomlSectionVersion finds `version = "..."` inside one of the named
// sections. It is deliberately tiny rather than a full TOML parser.
func tomlSectionVersion(p *project.Project, file string, sections ...string) string {
	data, err := p.ReadFile(file)
	if err != nil {
		return ""
	}
	current := ""
	for _, line := range strings.Split(string(data), "\n") {
		if m := tomlSection.FindStringSubmatch(line); m != nil {
			current = strings.TrimSpace(m[1])
			continue
		}
		for _, want := range sections {
			if current == want {
				if m := tomlVersion.FindStringSubmatch(line); m != nil {
					return m[1]
				}
			}
		}
	}
	return ""
}
