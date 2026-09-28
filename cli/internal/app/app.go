// Package app wires Shipcheck's commands together.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"

	"github.com/spf13/cobra"

	"github.com/App-Chef/shipcheck/cli/internal/output"
)

// Exit codes.
const (
	ExitReady    = 0 // every check passed or only warned
	ExitNotReady = 1 // at least one check failed
	ExitError    = 2 // Shipcheck itself could not run
)

// errNotReady signals exit code 1 after the report has been printed.
var errNotReady = errors.New("not ready to ship")

// options holds the global flags.
type options struct {
	verbose    bool
	json       bool
	noColor    bool
	configPath string
	dir        string
}

// Execute runs the CLI with args and returns the process exit code.
func Execute(args []string, stdout, stderr io.Writer, build BuildInfo) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	opts := &options{}
	root := newRootCommand(opts, stdout, build)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	err := root.ExecuteContext(ctx)
	switch {
	case err == nil:
		return ExitReady
	case errors.Is(err, errNotReady):
		return ExitNotReady
	}
	// Flag parsing can fail before opts.json is set; JSON consumers still
	// need a JSON error.
	if opts.json || slices.Contains(args, "--json") {
		output.WriteJSONError(stdout, err)
	} else {
		fmt.Fprintf(stderr, "shipcheck: %v\n", err)
	}
	return ExitError
}

func newRootCommand(opts *options, stdout io.Writer, build BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:   "shipcheck",
		Short: "Know you're ready before you ship.",
		Long: `Know you're ready before you ship.

Shipcheck checks your project for common release blockers:
uncommitted changes, failing tests, broken builds, missing
environment variables, missing README or license, and more.

Running "shipcheck" with no command is the same as "shipcheck check".
It exits 0 when ready, 1 when a check fails and 2 on errors.`,
		Example: `  shipcheck                        Check the current project
  shipcheck --verbose              Show commands, output and timings
  shipcheck --json                 Machine-readable report for CI
  shipcheck init                   Create a .shipcheck.yml
  shipcheck && ./deploy.sh         Deploy only when ready`,
		Version:       build.Version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCheck(cmd.Context(), opts, stdout, build)
		},
	}
	root.SetVersionTemplate("shipcheck {{.Version}}\n")
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w (run 'shipcheck help' for usage)", err)
	})

	flags := root.PersistentFlags()
	flags.BoolVar(&opts.verbose, "verbose", false, "show commands, output and timings")
	flags.BoolVar(&opts.json, "json", false, "print a JSON report instead of text")
	flags.StringVarP(&opts.configPath, "config", "c", "", "path to a config file (default .shipcheck.yml)")
	flags.StringVarP(&opts.dir, "dir", "C", ".", "project directory to check")
	flags.BoolVar(&opts.noColor, "no-color", false, "disable colored output (also honours NO_COLOR)")
	root.Flags().Bool("version", false, "print the version and exit")

	root.AddCommand(
		newCheckCommand(opts, stdout, build),
		newInitCommand(opts, stdout),
		newVersionCommand(opts, stdout, build),
	)
	return root
}

func newCheckCommand(opts *options, stdout io.Writer, build BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check whether the project is ready to ship (default)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCheck(cmd.Context(), opts, stdout, build)
		},
	}
}
