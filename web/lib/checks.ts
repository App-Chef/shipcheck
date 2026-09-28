export type Outcome = {
  status: "pass" | "warn" | "fail" | "skip";
  message: string;
};

export type CheckDoc = {
  /** Check ID as it appears in --json output. */
  id: string;
  /** Key used under `checks:` and `severity:` in .shipcheck.yml. */
  key: string;
  name: string;
  what: string;
  why: string;
  outcomes: Outcome[];
  config?: string;
  notes?: string[];
};

export type CheckGroup = {
  slug: string;
  label: string;
  title: string;
  summary: string;
  example: string;
  checks: CheckDoc[];
};

/**
 * The checks Shipcheck ships with. Messages are copied from the CLI
 * source (cli/internal/checks) and must be kept in sync with it.
 */
export const checkGroups: CheckGroup[] = [
  {
    slug: "git",
    label: "Git",
    title: "Git",
    summary: "Repository present, working tree clean, on the branch you ship from.",
    example: "✓ Working tree clean",
    checks: [
      {
        id: "git.repo",
        key: "git_repo",
        name: "Git repository",
        what: "Checks that the project is inside a Git repository. Works from any subdirectory of a repository, which makes monorepos straightforward.",
        why: "A release that is not in version control cannot be traced, reviewed or rolled back.",
        outcomes: [
          { status: "pass", message: "Git repository" },
          { status: "fail", message: "Not a Git repository" },
        ],
      },
      {
        id: "git.clean",
        key: "git_clean",
        name: "Working tree",
        what: "Runs git status and fails if there are uncommitted or untracked changes. Files ignored by .gitignore are ignored here too. --verbose lists the first ten changed paths.",
        why: "If what you deploy differs from what is committed, the release cannot be reproduced and the fix you forgot to commit will be lost.",
        outcomes: [
          { status: "pass", message: "Working tree clean" },
          { status: "fail", message: "3 uncommitted changes" },
          { status: "skip", message: "Working tree not checked (not a Git repository)" },
        ],
      },
      {
        id: "git.branch",
        key: "git_branch",
        name: "Branch",
        what: "Warns when you are not on a branch you ship from. Those branches come from git.protected_branches. Without configuration, Shipcheck uses main or master if either exists in the repository, and skips the check otherwise.",
        why: "Shipping from a feature branch is a common way to release code that was never reviewed or merged.",
        outcomes: [
          { status: "pass", message: "On branch main" },
          { status: "warn", message: "On branch feature/login, not main" },
          { status: "skip", message: "Detached HEAD (no branch to check)" },
        ],
        config: `git:
  protected_branches:
    - main
    - release`,
      },
    ],
  },
  {
    slug: "tests",
    label: "Tests",
    title: "Tests",
    summary: "Detects your test runner and runs it. Fails if any test fails.",
    example: "✓ Tests passed",
    checks: [
      {
        id: "tests",
        key: "tests",
        name: "Tests",
        what: "Detects and runs the project's test command with CI=true, so watch-mode runners exit after one run. Output is captured, not streamed. On failure, --verbose shows the last 20 lines with secret values redacted.",
        why: "“Tests pass on my machine” is only true if you actually ran them on the code you are about to ship.",
        outcomes: [
          { status: "pass", message: "Tests passed" },
          { status: "fail", message: "Tests failed (npm test)" },
          { status: "fail", message: "Tests not run: cargo is not installed" },
          { status: "skip", message: "No test command detected" },
        ],
        config: `commands:
  tests: npm run test:ci
  timeout: 10m`,
        notes: [
          "Go: go test ./... (go.mod)",
          "Node.js: npm test, pnpm run test, yarn run test or bun run test, chosen from packageManager or the lockfile. The default “no test specified” script is ignored.",
          "Python: pytest or python -m pytest, when pytest.ini, conftest.py, a tests/ directory or [tool.pytest] exists",
          "Java: ./mvnw test or mvn test (pom.xml), ./gradlew test or gradle test (build.gradle)",
          "PHP: php artisan test (Laravel), composer test, or vendor/bin/phpunit",
          "Rust: cargo test (Cargo.toml)",
        ],
      },
    ],
  },
  {
    slug: "build",
    label: "Build",
    title: "Production build",
    summary: "Runs your production build so compile errors surface before deploy.",
    example: "✓ Production build passed",
    checks: [
      {
        id: "build",
        key: "build",
        name: "Production build",
        what: "Detects and runs the project's production build command. The build writes its usual output (for example dist/ or .next/) and nothing else.",
        why: "Many errors only appear in production mode: type errors, missing imports or environment-specific config.",
        outcomes: [
          { status: "pass", message: "Production build passed" },
          { status: "fail", message: "Production build failed (npm run build)" },
          { status: "fail", message: "npm run build timed out after 10m0s" },
          { status: "skip", message: "No build command detected" },
        ],
        config: `commands:
  build:
    - npm run build
    - go build ./...`,
        notes: [
          "Go: go build ./...",
          "Node.js: the build script in package.json, run with your package manager",
          "Java: mvn package -DskipTests or gradle assemble (wrappers preferred)",
          "Rust: cargo build --release",
          "Python and PHP have no standard build step. Set commands.build if you have one.",
        ],
      },
    ],
  },
  {
    slug: "environment",
    label: "Environment",
    title: "Environment",
    summary: "Required variables are set, and values are never printed.",
    example: "✓ Environment checked",
    checks: [
      {
        id: "environment",
        key: "environment",
        name: "Environment",
        what: "Reads variable names from .env.example, .env.sample or .example.env, plus environment.required. Each one must have a non-empty value, either in the process environment or in a dotenv file (.env by default). The check also warns when a real .env file is committed to Git.",
        why: "A missing API key is the classic “works locally, breaks in production” bug. A committed .env is a leaked secret.",
        outcomes: [
          { status: "pass", message: "Environment checked" },
          { status: "fail", message: "2 environment variables missing" },
          { status: "warn", message: ".env is committed to Git" },
          { status: "skip", message: "No environment variables required" },
        ],
        config: `environment:
  required:
    - DATABASE_URL
    - API_KEY
  files:
    - .env
    - .env.local`,
        notes: [
          "Only variable names are ever printed. Values stay private.",
          "Values of required variables are replaced with **** in any test or build output Shipcheck shows.",
        ],
      },
    ],
  },
  {
    slug: "readme",
    label: "README",
    title: "README",
    summary: "The project explains what it is and how to run it.",
    example: "✓ README found",
    checks: [
      {
        id: "readme",
        key: "readme",
        name: "README",
        what: "Looks for README.md, case-insensitively. README, README.markdown, README.rst and README.txt also count. An empty README is reported as a warning.",
        why: "The README is the first thing users, teammates and your future self read.",
        outcomes: [
          { status: "pass", message: "README found" },
          { status: "warn", message: "No README" },
          { status: "warn", message: "README.md is empty" },
        ],
      },
    ],
  },
  {
    slug: "license",
    label: "License",
    title: "License",
    summary: "A license file exists, and common licenses are recognised.",
    example: "✓ License found (MIT)",
    checks: [
      {
        id: "license",
        key: "license",
        name: "License",
        what: "Looks for LICENSE, LICENSE.md or LICENSE.txt (also LICENCE and COPYING), and names MIT, Apache-2.0, GPL, AGPL, LGPL, MPL-2.0, ISC, BSD and the Unlicense when it recognises them.",
        why: "Without a license, others cannot legally use your code. For private projects, turn the check off.",
        outcomes: [
          { status: "pass", message: "License found (MIT)" },
          { status: "warn", message: "No license file" },
        ],
        config: `checks:
  license: false`,
      },
    ],
  },
  {
    slug: "version",
    label: "Version",
    title: "Version",
    summary: "Finds your version, and warns if that version was already released.",
    example: "✓ Version 1.4.0",
    checks: [
      {
        id: "version",
        key: "version",
        name: "Version",
        what: "Reads the version from package.json, Cargo.toml, pyproject.toml, composer.json, pom.xml, build.gradle(.kts) or a VERSION file, falling back to the latest Git tag. If a tag for that version (1.4.0 or v1.4.0) already exists at an older commit than HEAD, it warns.",
        why: "Shipping new code under an old version number makes bug reports and rollbacks ambiguous.",
        outcomes: [
          { status: "pass", message: "Version 1.4.0" },
          { status: "warn", message: "Version 1.4.0 was already released (v1.4.0)" },
          { status: "warn", message: "No version found" },
        ],
      },
    ],
  },
  {
    slug: "ci",
    label: "CI",
    title: "CI",
    summary: "Continuous integration is configured. Recommended, not required.",
    example: "✓ CI configured (GitHub Actions)",
    checks: [
      {
        id: "ci",
        key: "ci",
        name: "CI",
        what: "Detects GitHub Actions (.github/workflows/*.yml), GitLab CI (.gitlab-ci.yml), CircleCI (.circleci/), Jenkins (Jenkinsfile), Bitbucket Pipelines and Azure Pipelines. A missing CI setup is a warning, never a failure, unless you raise its severity.",
        why: "CI runs your checks on every push, not just when you remember.",
        outcomes: [
          { status: "pass", message: "CI configured (GitHub Actions)" },
          { status: "warn", message: "No CI configuration" },
        ],
        config: `severity:
  ci: error   # require CI`,
      },
    ],
  },
];

export const allChecks: CheckDoc[] = checkGroups.flatMap((g) => g.checks);

export const statusSymbol: Record<Outcome["status"], string> = {
  pass: "✓",
  warn: "⚠",
  fail: "✗",
  skip: "–",
};
