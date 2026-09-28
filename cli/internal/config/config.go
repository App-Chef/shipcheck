// Package config loads and validates .shipcheck.yml.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FileNames are the config file names Shipcheck looks for, in order.
var FileNames = []string{".shipcheck.yml", ".shipcheck.yaml"}

// DefaultTimeout bounds each test or build command.
const DefaultTimeout = 10 * time.Minute

// Severity overrides how a check's problems are reported.
type Severity string

const (
	// SeverityDefault keeps the check's own judgement.
	SeverityDefault Severity = ""
	// SeverityError turns warnings into failures.
	SeverityError Severity = "error"
	// SeverityWarn turns failures into warnings.
	SeverityWarn Severity = "warn"
)

// Config is the parsed contents of .shipcheck.yml.
type Config struct {
	Version     int                 `yaml:"version"`
	Checks      map[string]bool     `yaml:"checks"`
	Severity    map[string]Severity `yaml:"severity"`
	Git         GitConfig           `yaml:"git"`
	Environment EnvironmentConfig   `yaml:"environment"`
	Commands    CommandsConfig      `yaml:"commands"`

	// Path is the file the config was loaded from, or "" for defaults.
	Path string `yaml:"-"`
}

// GitConfig configures the git checks.
type GitConfig struct {
	// ProtectedBranches are the branches you ship from. When set,
	// Shipcheck warns if you are on any other branch.
	ProtectedBranches []string `yaml:"protected_branches"`
}

// EnvironmentConfig configures the environment check.
type EnvironmentConfig struct {
	// Required lists variable names that must be set.
	Required []string `yaml:"required"`
	// Files are dotenv files (relative to the project) that may provide
	// values. Defaults to [".env"].
	Files []string `yaml:"files"`
}

// CommandsConfig overrides the detected test and build commands.
type CommandsConfig struct {
	Tests   CommandList `yaml:"tests"`
	Build   CommandList `yaml:"build"`
	Timeout Duration    `yaml:"timeout"`
}

// CommandList accepts either a single command string or a list of them.
type CommandList []string

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *CommandList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var s string
		if err := node.Decode(&s); err != nil {
			return err
		}
		*c = CommandList{s}
		return nil
	case yaml.SequenceNode:
		var list []string
		if err := node.Decode(&list); err != nil {
			return err
		}
		*c = list
		return nil
	}
	return fmt.Errorf("line %d: expected a command string or a list of commands", node.Line)
}

// Duration is a time.Duration written as "90s" or "10m" in YAML.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("line %d: invalid duration %q (use values like 90s or 10m)", node.Line, s)
	}
	if parsed <= 0 {
		return fmt.Errorf("line %d: duration must be positive", node.Line)
	}
	*d = Duration(parsed)
	return nil
}

// UnmarshalYAML accepts "error", "warn" and "warning".
func (s *Severity) UnmarshalYAML(node *yaml.Node) error {
	var raw string
	if err := node.Decode(&raw); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "error", "fail":
		*s = SeverityError
	case "warn", "warning":
		*s = SeverityWarn
	default:
		return fmt.Errorf("line %d: invalid severity %q (use error or warn)", node.Line, raw)
	}
	return nil
}

// Default returns the configuration used when no file exists.
func Default() *Config {
	return &Config{Version: 1}
}

// Find returns the path of the config file in dir, or "" if none exists.
func Find(dir string) string {
	for _, name := range FileNames {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return path
		}
	}
	return ""
}

// Load reads and validates the config at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file %s does not exist", path)
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	cfg.Path = path
	return cfg, nil
}

// Parse decodes and validates config data. Unknown keys are rejected so
// typos never silently disable a check.
func Parse(data []byte) (*Config, error) {
	cfg := Default()
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, cleanYAMLError(err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (c *Config) validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d (expected 1)", c.Version)
	}
	for _, name := range c.Environment.Required {
		if !envNamePattern.MatchString(name) {
			return fmt.Errorf("environment.required: %q is not a valid variable name", name)
		}
	}
	for _, f := range c.Environment.Files {
		if err := validateRelPath(f); err != nil {
			return fmt.Errorf("environment.files: %w", err)
		}
	}
	for _, b := range c.Git.ProtectedBranches {
		if strings.TrimSpace(b) == "" {
			return errors.New("git.protected_branches: branch names cannot be empty")
		}
	}
	for field, list := range map[string]CommandList{"commands.tests": c.Commands.Tests, "commands.build": c.Commands.Build} {
		for _, cmd := range list {
			if _, err := ParseCommand(cmd); err != nil {
				return fmt.Errorf("%s: %w", field, err)
			}
		}
	}
	return nil
}

// ValidateCheckKeys ensures every key in checks and severity names a real
// check. known holds check IDs such as "git.clean".
func (c *Config) ValidateCheckKeys(known []string) error {
	valid := make(map[string]bool, len(known))
	keys := make([]string, 0, len(known))
	for _, id := range known {
		valid[Key(id)] = true
		keys = append(keys, Key(id))
	}
	sort.Strings(keys)
	check := func(section, key string) error {
		if valid[Key(key)] {
			return nil
		}
		return fmt.Errorf("%s: unknown check %q (valid checks: %s)", section, key, strings.Join(keys, ", "))
	}
	for key := range c.Checks {
		if err := check("checks", key); err != nil {
			return err
		}
	}
	for key := range c.Severity {
		if err := check("severity", key); err != nil {
			return err
		}
	}
	return nil
}

// Key converts a check ID ("git.clean") to its config key ("git_clean").
func Key(id string) string {
	return strings.ReplaceAll(id, ".", "_")
}

// Enabled reports whether the check with the given ID should run.
// Checks are enabled unless explicitly set to false.
func (c *Config) Enabled(id string) bool {
	for key, on := range c.Checks {
		if Key(key) == Key(id) {
			return on
		}
	}
	return true
}

// SeverityFor returns the configured severity for a check ID.
func (c *Config) SeverityFor(id string) Severity {
	for key, sev := range c.Severity {
		if Key(key) == Key(id) {
			return sev
		}
	}
	return SeverityDefault
}

// Timeout returns the per-command timeout.
func (c *Config) Timeout() time.Duration {
	if c.Commands.Timeout > 0 {
		return time.Duration(c.Commands.Timeout)
	}
	return DefaultTimeout
}

// EnvFiles returns the dotenv files to read values from.
func (c *Config) EnvFiles() []string {
	if len(c.Environment.Files) > 0 {
		return c.Environment.Files
	}
	return []string{".env"}
}

// validateRelPath rejects absolute paths and paths escaping the project.
func validateRelPath(p string) error {
	if p == "" {
		return errors.New("path cannot be empty")
	}
	if filepath.IsAbs(p) || strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) {
		return fmt.Errorf("%q must be relative to the project", p)
	}
	clean := filepath.ToSlash(filepath.Clean(p))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("%q must stay inside the project", p)
	}
	return nil
}

var typeSuffix = regexp.MustCompile(` in type config\.\w+`)

func cleanYAMLError(err error) error {
	msg := strings.TrimPrefix(err.Error(), "yaml: ")
	msg = strings.ReplaceAll(msg, "unmarshal errors:\n  ", "")
	msg = typeSuffix.ReplaceAllString(msg, "")
	return errors.New(msg)
}
