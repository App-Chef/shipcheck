package checks

import (
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

func TestReadme(t *testing.T) {
	expect(t, run(t, Readme, newEnv(t, testutil.Dir(t, nil), "", nil)), Warn, "No README")
	expect(t, run(t, Readme, newEnv(t, testutil.Dir(t, map[string]string{"README.md": "  \n"}), "", nil)), Warn, "README.md is empty")
	expect(t, run(t, Readme, newEnv(t, testutil.Dir(t, map[string]string{"README.md": "# App"}), "", nil)), Pass, "README found")
	expect(t, run(t, Readme, newEnv(t, testutil.Dir(t, map[string]string{"readme.md": "# App"}), "", nil)), Pass, "README found")
	expect(t, run(t, Readme, newEnv(t, testutil.Dir(t, map[string]string{"README.rst": "App"}), "", nil)), Pass, "README found")
	// A directory named README is not a README.
	expect(t, run(t, Readme, newEnv(t, testutil.Dir(t, map[string]string{"README/x": "x"}), "", nil)), Warn, "No README")
}

func TestLicense(t *testing.T) {
	expect(t, run(t, License, newEnv(t, testutil.Dir(t, nil), "", nil)), Warn, "No license file")

	cases := map[string]struct{ file, content, want string }{
		"mit":       {"LICENSE", "MIT License\n\nCopyright (c) 2026", "License found (MIT)"},
		"apache md": {"LICENSE.md", "Apache License\nVersion 2.0, January 2004", "License found (Apache-2.0)"},
		"gpl txt":   {"LICENSE.txt", "GNU GENERAL PUBLIC LICENSE\nVersion 3", "License found (GPL)"},
		"agpl":      {"LICENSE", "GNU AFFERO GENERAL PUBLIC LICENSE", "License found (AGPL)"},
		"lowercase": {"license", "ISC License", "License found (ISC)"},
		"british":   {"LICENCE", "Mozilla Public License Version 2.0", "License found (MPL-2.0)"},
		"unknown":   {"LICENSE", "All rights reserved.", "License found"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := testutil.Dir(t, map[string]string{tc.file: tc.content})
			res := run(t, License, newEnv(t, dir, "", nil))
			expect(t, res, Pass, tc.want)
			if name == "unknown" && res.Message != "License found" {
				t.Errorf("message = %q", res.Message)
			}
		})
	}
}

func TestCI(t *testing.T) {
	expect(t, run(t, CI, newEnv(t, testutil.Dir(t, nil), "", nil)), Warn, "No CI configuration")
	// An empty workflows directory is not CI.
	expect(t, run(t, CI, newEnv(t, testutil.Dir(t, map[string]string{".github/workflows/README.md": ""}), "", nil)), Warn, "No CI configuration")

	for file, want := range map[string]string{
		".github/workflows/ci.yml":  "GitHub Actions",
		".github/workflows/ci.yaml": "GitHub Actions",
		".gitlab-ci.yml":            "GitLab CI",
		".circleci/config.yml":      "CircleCI",
		"Jenkinsfile":               "Jenkins",
		"bitbucket-pipelines.yml":   "Bitbucket Pipelines",
		"azure-pipelines.yml":       "Azure Pipelines",
	} {
		dir := testutil.Dir(t, map[string]string{file: "x"})
		expect(t, run(t, CI, newEnv(t, dir, "", nil)), Pass, "CI configured ("+want+")")
	}

	dir := testutil.Dir(t, map[string]string{".github/workflows/ci.yml": "", "Jenkinsfile": ""})
	expect(t, run(t, CI, newEnv(t, dir, "", nil)), Pass, "GitHub Actions, Jenkins")
}

func TestVersion(t *testing.T) {
	expect(t, run(t, Version, newEnv(t, testutil.Dir(t, nil), "", nil)), Warn, "No version found")

	dir := testutil.Dir(t, map[string]string{"package.json": `{"version":"1.2.0"}`})
	res := run(t, Version, newEnv(t, dir, "", nil))
	expect(t, res, Pass, "Version 1.2.0")
	if len(res.Details) != 1 || res.Details[0] != "from package.json" {
		t.Errorf("details = %q", res.Details)
	}
}

func TestVersionAgainstGitTags(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"package.json": `{"version":"1.2.0"}`})
	testutil.GitRepo(t, dir)
	testutil.Git(t, dir, "tag", "v1.2.0")

	// HEAD is the tagged commit: this is the release.
	expect(t, run(t, Version, newEnv(t, dir, "", nil)), Pass, "Version 1.2.0")

	// New commits on top of an already released version.
	testutil.Commit(t, dir, "more work")
	res := run(t, Version, newEnv(t, dir, "", nil))
	expect(t, res, Warn, "Version 1.2.0 was already released (v1.2.0)")
	if res.Suggestion == "" {
		t.Error("expected a suggestion to bump the version")
	}
}

func TestVersionFromGitTag(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{"main.go": "package main"})
	testutil.GitRepo(t, dir)
	expect(t, run(t, Version, newEnv(t, dir, "", nil)), Warn, "No version found")

	testutil.Git(t, dir, "tag", "v0.4.0")
	testutil.Commit(t, dir, "next")
	res := run(t, Version, newEnv(t, dir, "", nil))
	expect(t, res, Pass, "Version v0.4.0")
}
