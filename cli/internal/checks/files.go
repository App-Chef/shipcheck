package checks

import (
	"context"
	"strings"
)

// Readme checks that the project explains itself.
var Readme = Check{
	ID:          "readme",
	Name:        "README",
	Description: "Checks for a README.md with content.",
	Run: func(_ context.Context, env *Env) Result {
		name := env.Project.FindFile("README.md", "README", "README.markdown", "README.rst", "README.txt")
		if name == "" {
			return Result{
				Status:     Warn,
				Message:    "No README",
				Suggestion: "Add a README.md that explains what the project does and how to run it.",
			}
		}
		data, err := env.Project.ReadFile(name)
		if err == nil && strings.TrimSpace(string(data)) == "" {
			return Result{
				Status:     Warn,
				Message:    name + " is empty",
				Suggestion: "Explain what the project does and how to run it.",
			}
		}
		return pass("README found", name)
	},
}

// License checks for a license file.
var License = Check{
	ID:          "license",
	Name:        "License",
	Description: "Checks for a LICENSE, LICENSE.md or LICENSE.txt file.",
	Run: func(_ context.Context, env *Env) Result {
		name := env.Project.FindFile("LICENSE", "LICENSE.md", "LICENSE.txt", "LICENCE", "LICENCE.md", "LICENCE.txt", "COPYING")
		if name == "" {
			return Result{
				Status:     Warn,
				Message:    "No license file",
				Suggestion: "Add a LICENSE file so others know how they may use the project (see choosealicense.com).",
			}
		}
		data, _ := env.Project.ReadFile(name)
		if kind := licenseKind(string(data)); kind != "" {
			return pass("License found ("+kind+")", name)
		}
		return pass("License found", name)
	},
}

// licenseKind recognises the most common licenses by their headings.
func licenseKind(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "mit license") || strings.Contains(t, "permission is hereby granted, free of charge"):
		return "MIT"
	case strings.Contains(t, "apache license") && strings.Contains(t, "version 2.0"):
		return "Apache-2.0"
	case strings.Contains(t, "gnu affero general public license"):
		return "AGPL"
	case strings.Contains(t, "gnu lesser general public license"):
		return "LGPL"
	case strings.Contains(t, "gnu general public license"):
		return "GPL"
	case strings.Contains(t, "mozilla public license"):
		return "MPL-2.0"
	case strings.Contains(t, "isc license"):
		return "ISC"
	case strings.Contains(t, "bsd") && strings.Contains(t, "redistribution and use in source and binary forms"):
		return "BSD"
	case strings.Contains(t, "this is free and unencumbered software released into the public domain"):
		return "Unlicense"
	}
	return ""
}

// ciProvider is a CI system and the path that indicates it.
type ciProvider struct {
	name string
	path string
	dir  bool
}

var ciProviders = []ciProvider{
	{"GitHub Actions", ".github/workflows", true},
	{"GitLab CI", ".gitlab-ci.yml", false},
	{"CircleCI", ".circleci", true},
	{"Jenkins", "Jenkinsfile", false},
	{"Bitbucket Pipelines", "bitbucket-pipelines.yml", false},
	{"Azure Pipelines", "azure-pipelines.yml", false},
}

// CI checks for continuous integration configuration. It warns rather
// than fails: CI is recommended, not required.
var CI = Check{
	ID:          "ci",
	Name:        "CI",
	Description: "Detects CI configuration for GitHub Actions, GitLab CI, CircleCI, Jenkins, Bitbucket or Azure Pipelines.",
	Run: func(_ context.Context, env *Env) Result {
		var found []string
		for _, p := range ciProviders {
			switch {
			case p.path == ".github/workflows":
				if env.Project.HasFileWithExt(p.path, ".yml", ".yaml") {
					found = append(found, p.name)
				}
			case p.dir && env.Project.IsDir(p.path), !p.dir && env.Project.IsFile(p.path):
				found = append(found, p.name)
			}
		}
		if len(found) == 0 {
			return Result{
				Status:     Warn,
				Message:    "No CI configuration",
				Suggestion: "Run tests on every push with CI, e.g. `shipcheck --json` in GitHub Actions.",
			}
		}
		return pass("CI configured ("+strings.Join(found, ", ")+")", found...)
	},
}
