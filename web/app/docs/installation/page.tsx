import type { Metadata } from "next";
import Link from "next/link";
import { CodeBlock } from "@/components/code-block";
import { DocPage, H2, Note } from "@/components/docs/doc-page";
import { site } from "@/lib/site";

export const metadata: Metadata = {
  title: "Installation",
  description: "Install the Shipcheck CLI with go install or build it from source.",
};

export default function Installation() {
  return (
    <DocPage
      href="/docs/installation"
      eyebrow="Getting started"
      title="Installation"
      lead="Shipcheck is a single Go binary with no runtime dependencies. It runs on Linux, macOS and Windows."
    >
      <H2 id="go-install">With Go</H2>
      <p>
        If you have Go {site.minGo} or newer, install the latest version with one command:
      </p>
      <CodeBlock lang="bash" code={`$ ${site.installCommand}`} />
      <p>
        The binary is placed in <code>$(go env GOPATH)/bin</code>, which is usually <code>~/go/bin</code>. Make sure
        that directory is on your <code>PATH</code>:
      </p>
      <CodeBlock
        lang="bash"
        title="~/.bashrc or ~/.zshrc"
        code={`export PATH="$PATH:$(go env GOPATH)/bin"`}
      />
      <p>
        On Windows, Go adds <code>%USERPROFILE%\go\bin</code> to your <code>PATH</code> during installation.
      </p>

      <H2 id="from-source">From source</H2>
      <p>Clone the repository and build the binary yourself:</p>
      <CodeBlock
        lang="bash"
        code={`$ git clone ${site.repo}.git
$ cd shipcheck/cli
$ go build -o shipcheck ./cmd/shipcheck
$ ./shipcheck version`}
      />
      <p>Move the resulting <code>shipcheck</code> binary anywhere on your <code>PATH</code>.</p>

      <H2 id="verify">Verify the installation</H2>
      <CodeBlock
        lang="bash"
        code={`$ shipcheck --version`}
      />
      <p>
        This prints the installed version. Builds installed with <code>go install</code> report the module version,
        and <code>shipcheck version --verbose</code> also shows the commit, build date and Go version.
      </p>

      <H2 id="requirements">Requirements</H2>
      <ul>
        <li>
          <strong>Git</strong> for the Git checks. Without it, the working tree and branch checks are skipped and
          everything else still runs.
        </li>
        <li>
          <strong>Your project&apos;s toolchain</strong> (for example <code>node</code>, <code>cargo</code> or{" "}
          <code>mvn</code>) for the tests and build checks. Shipcheck runs your commands; it doesn&apos;t bundle them.
        </li>
      </ul>

      <Note title="Other install methods">
        <p>
          Prebuilt binaries and package managers (such as Homebrew) are not available yet. Until they are, use{" "}
          <code>go install</code> or build from source.
        </p>
      </Note>

      <H2 id="uninstall">Uninstall</H2>
      <p>Shipcheck stores no data outside your project. Delete the binary to remove it:</p>
      <CodeBlock lang="bash" code={`$ rm "$(go env GOPATH)/bin/shipcheck"`} />
      <p>
        Next: <Link href="/docs/quick-start">run your first check</Link>.
      </p>
    </DocPage>
  );
}
