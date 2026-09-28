package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/App-Chef/shipcheck/cli/internal/checks"
	"github.com/App-Chef/shipcheck/cli/internal/config"
	"github.com/App-Chef/shipcheck/cli/internal/detector"
	"github.com/App-Chef/shipcheck/cli/internal/project"
)

func newInitCommand(opts *options, stdout io.Writer) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a .shipcheck.yml for this project",
		Long: `Create a commented .shipcheck.yml in the project directory,
pre-filled with the test and build commands Shipcheck detected.
Existing files are never overwritten unless --force is given.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runInit(opts, stdout, force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing .shipcheck.yml")
	return cmd
}

func runInit(opts *options, stdout io.Writer, force bool) error {
	proj, err := project.Open(opts.dir)
	if err != nil {
		return err
	}
	path := filepath.Join(proj.Root, config.FileNames[0])
	if existing := config.Find(proj.Root); existing != "" && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", filepath.Base(existing))
	}

	d := detector.Detect(proj)
	data := config.TemplateData{EnvTemplate: checks.NewEnv(proj, config.Default(), nil).EnvTemplate()}
	if proj.Git.Available() && proj.Git.IsRepo() {
		data.Branch, _ = proj.Git.Branch()
	}
	for _, c := range d.Tests {
		data.TestCommands = append(data.TestCommands, c.String())
	}
	for _, c := range d.Builds {
		data.BuildCommands = append(data.BuildCommands, c.String())
	}

	content := config.Template(data)
	// Never ship a template we cannot read back.
	if _, err := config.Parse(content); err != nil {
		return fmt.Errorf("generated config is invalid: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}

	if opts.json {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{"created": path})
	}
	st := styleFor(stdout, opts.noColor)
	fmt.Fprintf(stdout, "%s Created %s\n", st.Green("✓"), config.FileNames[0])
	fmt.Fprintln(stdout, st.Dim("  Edit it to fit your project, then run `shipcheck`."))
	return nil
}
