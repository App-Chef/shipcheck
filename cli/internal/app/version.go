package app

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// BuildInfo describes the binary. Release builds set it with -ldflags.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// ResolveBuildInfo fills gaps from the Go module information embedded by
// `go install`, so installed binaries report a real version.
func ResolveBuildInfo(b BuildInfo) BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return b
	}
	if (b.Version == "" || b.Version == "dev") && info.Main.Version != "" && info.Main.Version != "(devel)" {
		b.Version = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if b.Commit == "" {
				b.Commit = s.Value
			}
		case "vcs.time":
			if b.Date == "" {
				b.Date = s.Value
			}
		}
	}
	if len(b.Commit) > 12 {
		b.Commit = b.Commit[:12]
	}
	if b.Version == "" {
		b.Version = "dev"
	}
	return b
}

func newVersionCommand(opts *options, stdout io.Writer, build BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Shipcheck version",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if opts.json {
				enc := json.NewEncoder(stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]string{
					"version": build.Version,
					"commit":  build.Commit,
					"date":    build.Date,
					"go":      runtime.Version(),
					"os":      runtime.GOOS,
					"arch":    runtime.GOARCH,
				})
			}
			fmt.Fprintf(stdout, "shipcheck %s\n", build.Version)
			if opts.verbose {
				if build.Commit != "" {
					fmt.Fprintf(stdout, "commit  %s\n", build.Commit)
				}
				if build.Date != "" {
					fmt.Fprintf(stdout, "built   %s\n", build.Date)
				}
				fmt.Fprintf(stdout, "go      %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
			}
			return nil
		},
	}
}
