package checks

import (
	"context"
	"fmt"

	"github.com/App-Chef/shipcheck/cli/internal/detector"
)

// Version detects the project version and warns when it has already been
// released but the code has moved on.
var Version = Check{
	ID:          "version",
	Name:        "Version",
	Description: "Detects the project version from its manifest or Git tags, and warns when that version was already tagged at an older commit.",
	Run: func(_ context.Context, env *Env) Result {
		git := env.Project.Git
		useGit := git.Available() && git.IsRepo()

		if v, ok := detector.DetectVersion(env.Project); ok {
			if useGit {
				if tag, stale := staleTag(env, v.Value); stale {
					return Result{
						Status:     Warn,
						Message:    fmt.Sprintf("Version %s was already released (%s)", v.Value, tag),
						Suggestion: fmt.Sprintf("Bump the version in %s before shipping new changes.", v.Source),
						Details:    []string{"tag " + tag + " points to an older commit than HEAD"},
					}
				}
			}
			return pass(fmt.Sprintf("Version %s", v.Value), "from "+v.Source)
		}
		if useGit {
			if tag := git.LatestTag(); tag != "" {
				return pass("Version "+tag, "from git tag")
			}
		}
		return Result{
			Status:     Warn,
			Message:    "No version found",
			Suggestion: "Declare a version in your manifest or tag a release (git tag v0.1.0).",
		}
	},
}

// staleTag reports whether version is already tagged at a commit other
// than HEAD, meaning the next release would reuse an old version number.
func staleTag(env *Env, version string) (string, bool) {
	git := env.Project.Git
	head, err := git.RevParse("HEAD")
	if err != nil {
		return "", false
	}
	for _, tag := range []string{"v" + version, version} {
		commit, err := git.RevParse("refs/tags/" + tag)
		if err != nil || commit == "" {
			continue
		}
		return tag, commit != head
	}
	return "", false
}
