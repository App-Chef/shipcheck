// Package detector recognises project ecosystems and the commands they
// use to run tests and production builds.
package detector

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/App-Chef/shipcheck/cli/internal/project"
)

// Ecosystem identifies a language toolchain.
type Ecosystem string

const (
	Go     Ecosystem = "go"
	Node   Ecosystem = "node"
	Python Ecosystem = "python"
	Maven  Ecosystem = "maven"
	Gradle Ecosystem = "gradle"
	PHP    Ecosystem = "php"
	Rust   Ecosystem = "rust"
)

// Command is a program and its arguments, run without a shell.
type Command struct {
	Ecosystem Ecosystem
	Args      []string
}

// String renders the command for display.
func (c Command) String() string {
	return strings.Join(c.Args, " ")
}

// Detection is everything the detector found in a project.
type Detection struct {
	Ecosystems []Ecosystem
	Tests      []Command
	Builds     []Command
}

// Detect inspects the project root and returns the detected ecosystems,
// test commands and build commands, in a stable order.
func Detect(p *project.Project) Detection {
	var d Detection
	add := func(e Ecosystem, test, build []string) {
		d.Ecosystems = append(d.Ecosystems, e)
		if test != nil {
			d.Tests = append(d.Tests, Command{Ecosystem: e, Args: test})
		}
		if build != nil {
			d.Builds = append(d.Builds, Command{Ecosystem: e, Args: build})
		}
	}

	if p.IsFile("go.mod") {
		add(Go, []string{"go", "test", "./..."}, []string{"go", "build", "./..."})
	}
	if p.IsFile("package.json") {
		test, build := nodeCommands(p)
		add(Node, test, build)
	}
	if isPython(p) {
		add(Python, pythonTest(p), nil)
	}
	if p.IsFile("pom.xml") {
		mvn := wrapper(p, "mvnw", "mvn")
		add(Maven, []string{mvn, "test"}, []string{mvn, "package", "-DskipTests"})
	}
	if p.IsFile("build.gradle") || p.IsFile("build.gradle.kts") {
		gradle := wrapper(p, "gradlew", "gradle")
		add(Gradle, []string{gradle, "test"}, []string{gradle, "assemble"})
	}
	if p.IsFile("composer.json") {
		add(PHP, phpTest(p), nil)
	}
	if p.IsFile("Cargo.toml") {
		add(Rust, []string{"cargo", "test"}, []string{"cargo", "build", "--release"})
	}
	return d
}

// npmPlaceholder is the test script `npm init` writes by default.
var npmPlaceholder = regexp.MustCompile(`no test specified`)

type packageJSON struct {
	Version        string            `json:"version"`
	Scripts        map[string]string `json:"scripts"`
	PackageManager string            `json:"packageManager"`
}

func readPackageJSON(p *project.Project) (packageJSON, bool) {
	var pkg packageJSON
	data, err := p.ReadFile("package.json")
	if err != nil {
		return pkg, false
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return pkg, false
	}
	return pkg, true
}

func nodeCommands(p *project.Project) (test, build []string) {
	pkg, ok := readPackageJSON(p)
	if !ok {
		return nil, nil
	}
	pm := nodePackageManager(p, pkg.PackageManager)
	if script, ok := pkg.Scripts["test"]; ok && strings.TrimSpace(script) != "" && !npmPlaceholder.MatchString(script) {
		if pm == "npm" {
			test = []string{"npm", "test"}
		} else {
			// `bun test` runs Bun's own runner, so always use `run`.
			test = []string{pm, "run", "test"}
		}
	}
	if script, ok := pkg.Scripts["build"]; ok && strings.TrimSpace(script) != "" {
		build = []string{pm, "run", "build"}
	}
	return test, build
}

func nodePackageManager(p *project.Project, declared string) string {
	if name, _, _ := strings.Cut(declared, "@"); name != "" {
		switch name {
		case "npm", "pnpm", "yarn", "bun":
			return name
		}
	}
	switch {
	case p.IsFile("pnpm-lock.yaml"):
		return "pnpm"
	case p.IsFile("yarn.lock"):
		return "yarn"
	case p.IsFile("bun.lockb") || p.IsFile("bun.lock"):
		return "bun"
	}
	return "npm"
}

func isPython(p *project.Project) bool {
	for _, f := range []string{"pyproject.toml", "setup.py", "setup.cfg", "requirements.txt", "pytest.ini", "tox.ini", "Pipfile"} {
		if p.IsFile(f) {
			return true
		}
	}
	return false
}

// pythonTest returns a pytest command when the project has tests.
func pythonTest(p *project.Project) []string {
	hasTests := p.IsFile("pytest.ini") || p.IsFile("conftest.py") ||
		p.HasFileWithExt("tests", ".py") || p.HasFileWithExt("test", ".py")
	if !hasTests {
		if data, err := p.ReadFile("pyproject.toml"); err == nil && strings.Contains(string(data), "[tool.pytest") {
			hasTests = true
		}
	}
	if !hasTests {
		return nil
	}
	if _, err := exec.LookPath("pytest"); err == nil {
		return []string{"pytest"}
	}
	return []string{pythonBinary(), "-m", "pytest"}
}

func pythonBinary() string {
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return "python"
}

func phpTest(p *project.Project) []string {
	if p.IsFile("artisan") {
		return []string{"php", "artisan", "test"}
	}
	var composer struct {
		Scripts map[string]json.RawMessage `json:"scripts"`
	}
	if data, err := p.ReadFile("composer.json"); err == nil && json.Unmarshal(data, &composer) == nil {
		if _, ok := composer.Scripts["test"]; ok {
			return []string{"composer", "test"}
		}
	}
	if p.IsFile("phpunit.xml") || p.IsFile("phpunit.xml.dist") {
		return []string{"vendor/bin/phpunit"}
	}
	return nil
}

// wrapper prefers a project-local build wrapper (./mvnw, ./gradlew).
func wrapper(p *project.Project, script, fallback string) string {
	if runtime.GOOS == "windows" {
		for _, ext := range []string{".cmd", ".bat"} {
			if p.IsFile(script + ext) {
				return `.\` + script + ext
			}
		}
		return fallback
	}
	if p.IsFile(script) {
		return "./" + script
	}
	return fallback
}
