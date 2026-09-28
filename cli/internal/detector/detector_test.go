package detector

import (
	"runtime"
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/project"
	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

func open(t *testing.T, files map[string]string) *project.Project {
	t.Helper()
	p, err := project.Open(testutil.Dir(t, files))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func commands(cs []Command) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.String()
	}
	return out
}

func TestDetect(t *testing.T) {
	mvn, gradle := "./mvnw", "./gradlew"
	if runtime.GOOS == "windows" {
		mvn, gradle = `.\mvnw.cmd`, `.\gradlew.bat`
	}
	mvnWrapper, gradleWrapper := "mvnw", "gradlew"
	if runtime.GOOS == "windows" {
		mvnWrapper, gradleWrapper = "mvnw.cmd", "gradlew.bat"
	}

	tests := []struct {
		name   string
		files  map[string]string
		eco    []Ecosystem
		tests  []string
		builds []string
	}{
		{"empty", nil, nil, nil, nil},
		{
			"go",
			map[string]string{"go.mod": "module x\n"},
			[]Ecosystem{Go}, []string{"go test ./..."}, []string{"go build ./..."},
		},
		{
			"npm with scripts",
			map[string]string{"package.json": `{"scripts":{"test":"vitest run","build":"next build"}}`},
			[]Ecosystem{Node}, []string{"npm test"}, []string{"npm run build"},
		},
		{
			"npm placeholder test script is ignored",
			map[string]string{"package.json": `{"scripts":{"test":"echo \"Error: no test specified\" && exit 1"}}`},
			[]Ecosystem{Node}, nil, nil,
		},
		{
			"pnpm from lockfile",
			map[string]string{"package.json": `{"scripts":{"test":"jest","build":"tsc"}}`, "pnpm-lock.yaml": ""},
			[]Ecosystem{Node}, []string{"pnpm run test"}, []string{"pnpm run build"},
		},
		{
			"yarn from packageManager field",
			map[string]string{"package.json": `{"packageManager":"yarn@4.1.0","scripts":{"test":"jest"}}`},
			[]Ecosystem{Node}, []string{"yarn run test"}, nil,
		},
		{
			"bun uses run so the script runs",
			map[string]string{"package.json": `{"scripts":{"test":"vitest"}}`, "bun.lock": ""},
			[]Ecosystem{Node}, []string{"bun run test"}, nil,
		},
		{
			"invalid package.json",
			map[string]string{"package.json": `{nope`},
			[]Ecosystem{Node}, nil, nil,
		},
		{
			"python without tests",
			map[string]string{"requirements.txt": "flask\n"},
			[]Ecosystem{Python}, nil, nil,
		},
		{
			"maven with wrapper",
			map[string]string{"pom.xml": "<project/>", mvnWrapper: ""},
			[]Ecosystem{Maven}, []string{mvn + " test"}, []string{mvn + " package -DskipTests"},
		},
		{
			"maven without wrapper",
			map[string]string{"pom.xml": "<project/>"},
			[]Ecosystem{Maven}, []string{"mvn test"}, []string{"mvn package -DskipTests"},
		},
		{
			"gradle kotlin with wrapper",
			map[string]string{"build.gradle.kts": "", gradleWrapper: ""},
			[]Ecosystem{Gradle}, []string{gradle + " test"}, []string{gradle + " assemble"},
		},
		{
			"laravel",
			map[string]string{"composer.json": "{}", "artisan": ""},
			[]Ecosystem{PHP}, []string{"php artisan test"}, nil,
		},
		{
			"composer test script",
			map[string]string{"composer.json": `{"scripts":{"test":"phpunit"}}`},
			[]Ecosystem{PHP}, []string{"composer test"}, nil,
		},
		{
			"phpunit config",
			map[string]string{"composer.json": `{}`, "phpunit.xml.dist": ""},
			[]Ecosystem{PHP}, []string{"vendor/bin/phpunit"}, nil,
		},
		{
			"rust",
			map[string]string{"Cargo.toml": "[package]\n"},
			[]Ecosystem{Rust}, []string{"cargo test"}, []string{"cargo build --release"},
		},
		{
			"go and node together",
			map[string]string{"go.mod": "module x\n", "package.json": `{"scripts":{"build":"vite build"}}`},
			[]Ecosystem{Go, Node}, []string{"go test ./..."}, []string{"go build ./...", "npm run build"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := Detect(open(t, tc.files))
			if !equal(d.Ecosystems, tc.eco) {
				t.Errorf("ecosystems = %v, want %v", d.Ecosystems, tc.eco)
			}
			if got := commands(d.Tests); !equal(got, tc.tests) {
				t.Errorf("tests = %q, want %q", got, tc.tests)
			}
			if got := commands(d.Builds); !equal(got, tc.builds) {
				t.Errorf("builds = %q, want %q", got, tc.builds)
			}
		})
	}
}

func TestDetectPythonTests(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"pytest.ini":     {"pytest.ini": "[pytest]\n"},
		"tests dir":      {"pyproject.toml": "[project]\n", "tests/test_app.py": ""},
		"pyproject tool": {"pyproject.toml": "[tool.pytest.ini_options]\n"},
		"conftest":       {"setup.py": "", "conftest.py": ""},
	} {
		t.Run(name, func(t *testing.T) {
			d := Detect(open(t, files))
			if len(d.Tests) != 1 || d.Tests[0].Ecosystem != Python {
				t.Fatalf("tests = %v", d.Tests)
			}
			args := d.Tests[0].Args
			if args[0] != "pytest" && args[len(args)-1] != "pytest" {
				t.Errorf("unexpected python test command %q", d.Tests[0])
			}
		})
	}
}

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string]string
		want   string
		source string
	}{
		{"package.json", map[string]string{"package.json": `{"version":"1.4.2"}`}, "1.4.2", "package.json"},
		{"cargo", map[string]string{"Cargo.toml": "[dependencies]\nversion = \"9\"\n\n[package]\nname = \"x\"\nversion = \"0.3.0\"\n"}, "0.3.0", "Cargo.toml"},
		{"cargo workspace", map[string]string{"Cargo.toml": "[workspace.package]\nversion = \"2.0.0\"\n"}, "2.0.0", "Cargo.toml"},
		{"pyproject", map[string]string{"pyproject.toml": "[project]\nname = 'x'\nversion = '0.9.1'\n"}, "0.9.1", "pyproject.toml"},
		{"poetry", map[string]string{"pyproject.toml": "[tool.poetry]\nversion = \"1.0.0b1\"\n"}, "1.0.0b1", "pyproject.toml"},
		{"composer", map[string]string{"composer.json": `{"version":"3.1.0"}`}, "3.1.0", "composer.json"},
		{
			"pom ignores parent and dependencies",
			map[string]string{"pom.xml": `<project>
  <parent><version>9.9.9</version></parent>
  <dependencies><dependency><version>8.8.8</version></dependency></dependencies>
  <version>1.2.3</version>
</project>`},
			"1.2.3", "pom.xml",
		},
		{"gradle groovy", map[string]string{"build.gradle": "plugins { id 'java' }\nversion '0.5.0'\n"}, "0.5.0", "build.gradle"},
		{"gradle kts", map[string]string{"build.gradle.kts": "version = \"0.6.0\"\n"}, "0.6.0", "build.gradle.kts"},
		{"VERSION file", map[string]string{"VERSION": "2.1.0\n"}, "2.1.0", "VERSION"},
		{"package.json wins", map[string]string{"package.json": `{"version":"1.0.0"}`, "VERSION": "5.0.0"}, "1.0.0", "package.json"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := DetectVersion(open(t, tc.files))
			if !ok {
				t.Fatal("no version detected")
			}
			if v.Value != tc.want || v.Source != tc.source {
				t.Errorf("got %+v, want %s from %s", v, tc.want, tc.source)
			}
		})
	}

	for name, files := range map[string]map[string]string{
		"nothing":               nil,
		"package without field": {"package.json": `{"name":"x"}`},
		"pom property":          {"pom.xml": "<project><version>${revision}</version></project>"},
		"cargo dependency only": {"Cargo.toml": "[dependencies]\nversion = \"1\"\n"},
	} {
		t.Run("none/"+name, func(t *testing.T) {
			if v, ok := DetectVersion(open(t, files)); ok {
				t.Errorf("unexpected version %+v", v)
			}
		})
	}
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
