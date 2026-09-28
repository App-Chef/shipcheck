package checks

import (
	"reflect"
	"strings"
	"testing"

	"github.com/App-Chef/shipcheck/cli/internal/testutil"
)

func TestEnvironmentSkipsWithoutRequirements(t *testing.T) {
	dir := testutil.Dir(t, nil)
	expect(t, run(t, Environment, newEnv(t, dir, "", nil)), Skip, "No environment variables required")
}

func TestEnvironmentTemplateVariables(t *testing.T) {
	for _, tpl := range EnvTemplates {
		t.Run(tpl, func(t *testing.T) {
			dir := testutil.Dir(t, map[string]string{
				tpl:    "# Database\nSHIPCHECK_TEST_DB=postgres://localhost\nexport SHIPCHECK_TEST_KEY=\n",
				".env": "SHIPCHECK_TEST_DB=postgres://prod-host/db\n",
			})
			res := run(t, Environment, newEnv(t, dir, "", nil))
			expect(t, res, Fail, "1 environment variable missing")
			if !reflect.DeepEqual(res.Details, []string{"missing: SHIPCHECK_TEST_KEY"}) {
				t.Errorf("details = %q", res.Details)
			}
			assertNoValues(t, res, "postgres")

			t.Setenv("SHIPCHECK_TEST_KEY", "from-process-env")
			res = run(t, Environment, newEnv(t, dir, "", nil))
			expect(t, res, Pass, "Environment checked")
			assertNoValues(t, res, "from-process-env", "postgres")
		})
	}
}

func TestEnvironmentConfiguredRequirements(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{".env.local": "SHIPCHECK_A=1\n"})
	cfg := "version: 1\nenvironment:\n  required: [SHIPCHECK_A, SHIPCHECK_B]\n  files: [.env.local]\n"
	res := run(t, Environment, newEnv(t, dir, cfg, nil))
	expect(t, res, Fail, "1 environment variable missing")
	if !strings.Contains(res.Suggestion, "SHIPCHECK_B") || !strings.Contains(res.Suggestion, ".env.local") {
		t.Errorf("suggestion = %q", res.Suggestion)
	}

	t.Setenv("SHIPCHECK_B", "secret-value")
	res = run(t, Environment, newEnv(t, dir, cfg, nil))
	expect(t, res, Pass, "Environment checked")
	assertNoValues(t, res, "secret-value")
}

func TestEnvironmentEmptyValueCountsAsMissing(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{".env.example": "SHIPCHECK_EMPTY=\n", ".env": "SHIPCHECK_EMPTY=\"\"\n"})
	expect(t, run(t, Environment, newEnv(t, dir, "", nil)), Fail, "1 environment variable missing")
}

func TestEnvironmentTemplateWithNoVariables(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{".env.example": "# nothing yet\n"})
	expect(t, run(t, Environment, newEnv(t, dir, "", nil)), Pass, ".env.example lists no variables")
}

func TestEnvironmentWarnsOnCommittedDotenv(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{".env.example": "SHIPCHECK_T=\n", ".env": "SHIPCHECK_T=abc123\n"})
	testutil.GitRepo(t, dir)
	res := run(t, Environment, newEnv(t, dir, "", nil))
	expect(t, res, Warn, ".env is committed to Git")
	assertNoValues(t, res, "abc123")

	// Templates are meant to be committed.
	dir = testutil.Dir(t, map[string]string{".env.example": "SHIPCHECK_T=\n", ".gitignore": ".env\n"})
	testutil.GitRepo(t, dir)
	testutil.Write(t, dir, map[string]string{".env": "SHIPCHECK_T=abc123\n"})
	expect(t, run(t, Environment, newEnv(t, dir, "", nil)), Pass, "Environment checked")
}

func TestIsDotenvName(t *testing.T) {
	for name, want := range map[string]bool{
		".env": true, ".env.local": true, ".env.production": true,
		".env.example": false, ".env.sample": false, ".env.template": false,
		".envrc": false, "env": false, ".example.env": false,
	} {
		if got := isDotenvName(name); got != want {
			t.Errorf("isDotenvName(%q) = %v", name, got)
		}
	}
}

func TestParseDotenv(t *testing.T) {
	got := ParseDotenv(strings.Join([]string{
		"# comment",
		"",
		"PLAIN=value",
		"export EXPORTED=yes",
		`DOUBLE="quoted value"`,
		`SINGLE='it''s'`,
		"INLINE=value # comment",
		"HASH=abc#def",
		"SPACED = spaced ",
		"EMPTY=",
		"not a pair",
		"1BAD=x",
		"DOTTED.KEY=ok\r",
	}, "\n"))
	want := []KeyValue{
		{"PLAIN", "value"},
		{"EXPORTED", "yes"},
		{"DOUBLE", "quoted value"},
		{"SINGLE", "it"},
		{"INLINE", "value"},
		{"HASH", "abc#def"},
		{"SPACED", "spaced"},
		{"EMPTY", ""},
		{"DOTTED.KEY", "ok"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRedactIgnoresShortValues(t *testing.T) {
	dir := testutil.Dir(t, map[string]string{".env.example": "SHIPCHECK_R1=\nSHIPCHECK_R2=\n", ".env": "SHIPCHECK_R1=on\nSHIPCHECK_R2=hunter2\n"})
	env := newEnv(t, dir, "", nil)
	if got := env.Redact("on hunter2 done"); got != "on **** done" {
		t.Errorf("got %q", got)
	}
}

// assertNoValues fails if any secret appears anywhere in the result.
func assertNoValues(t *testing.T, r Result, secrets ...string) {
	t.Helper()
	all := r.Message + r.Suggestion + strings.Join(r.Details, "")
	for _, s := range secrets {
		if strings.Contains(all, s) {
			t.Errorf("result exposes %q: %+v", s, r)
		}
	}
}
