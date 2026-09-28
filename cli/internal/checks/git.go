package checks

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// GitRepo checks that the project is under version control.
var GitRepo = Check{
	ID:          "git.repo",
	Name:        "Git repository",
	Description: "Checks that the project is inside a Git repository.",
	Run: func(_ context.Context, env *Env) Result {
		if env.Project.Git.IsRepo() {
			return pass("Git repository")
		}
		return Result{
			Status:     Fail,
			Message:    "Not a Git repository",
			Suggestion: "Run `git init` and commit your project so every release is traceable.",
		}
	},
}

// GitClean checks for uncommitted changes.
var GitClean = Check{
	ID:          "git.clean",
	Name:        "Working tree",
	Description: "Checks that there are no uncommitted or untracked changes.",
	Run: func(_ context.Context, env *Env) Result {
		git := env.Project.Git
		if !git.Available() {
			return skip("Git is not installed")
		}
		if !git.IsRepo() {
			return skip("Working tree not checked (not a Git repository)")
		}
		changes, err := git.Changes()
		if err != nil {
			return Result{Status: Fail, Message: "Could not read Git status", Details: []string{err.Error()}}
		}
		if len(changes) == 0 {
			return pass("Working tree clean")
		}
		return Result{
			Status:     Fail,
			Message:    plural(len(changes), "uncommitted change", "uncommitted changes"),
			Suggestion: "Commit or stash your changes so you ship exactly what is in Git.",
			Details:    limitLines(changes, 10),
		}
	},
}

// defaultShipBranches are used when git.protected_branches is not set.
var defaultShipBranches = []string{"main", "master"}

// GitBranch checks that you are on a branch you ship from.
var GitBranch = Check{
	ID:          "git.branch",
	Name:        "Branch",
	Description: "Checks that you are on a branch you ship from (git.protected_branches, default main or master).",
	Run: func(_ context.Context, env *Env) Result {
		git := env.Project.Git
		if !git.Available() {
			return skip("Git is not installed")
		}
		if !git.IsRepo() {
			return skip("Branch not checked (not a Git repository)")
		}
		branch, err := git.Branch()
		if err != nil {
			return Result{Status: Fail, Message: "Could not read the current branch", Details: []string{err.Error()}}
		}
		if branch == "" {
			return skip("Detached HEAD (no branch to check)")
		}

		allowed := env.Config.Git.ProtectedBranches
		if len(allowed) == 0 {
			// Without configuration, only check when a conventional
			// release branch actually exists in this repository.
			for _, b := range defaultShipBranches {
				if b == branch {
					allowed = []string{b}
					break
				}
				if _, err := git.RevParse("refs/heads/" + b); err == nil {
					allowed = append(allowed, b)
				}
			}
			if len(allowed) == 0 {
				return skip("No release branch configured")
			}
		}
		if slices.Contains(allowed, branch) {
			return pass("On branch " + branch)
		}
		return Result{
			Status:     Warn,
			Message:    fmt.Sprintf("On branch %s, not %s", branch, orList(allowed)),
			Suggestion: "Merge into a release branch before shipping, or add this branch to git.protected_branches.",
		}
	},
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

func orList(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " or " + items[len(items)-1]
}

// limitLines keeps the first n lines and notes how many were dropped.
func limitLines(lines []string, n int) []string {
	if len(lines) <= n {
		return lines
	}
	out := append([]string(nil), lines[:n]...)
	return append(out, fmt.Sprintf("… and %d more", len(lines)-n))
}
