// Command shipcheck checks whether a project is ready to ship.
package main

import (
	"os"

	"github.com/App-Chef/shipcheck/cli/internal/app"
)

// Set by release builds:
//
//	go build -ldflags "-X main.version=v0.1.0 -X main.commit=$(git rev-parse HEAD)"
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	build := app.ResolveBuildInfo(app.BuildInfo{Version: version, Commit: commit, Date: date})
	os.Exit(app.Execute(os.Args[1:], os.Stdout, os.Stderr, build))
}
