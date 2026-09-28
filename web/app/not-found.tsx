import Link from "next/link";
import { Terminal } from "@/components/terminal";

export default function NotFound() {
  return (
    <div className="container-page flex flex-col items-start gap-8 py-20 sm:py-28">
      <div>
        <p className="eyebrow">404</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-[-0.03em] sm:text-5xl">Page not found</h1>
        <p className="mt-4 text-lg text-muted">This page doesn&apos;t exist, or it moved.</p>
      </div>
      <Terminal
        className="w-full max-w-xl"
        label="Page not found"
        title="~"
        lines={[
          { tone: "cmd", text: "$ cd this-page" },
          { tone: "plain", text: "cd: no such file or directory: this-page" },
        ]}
      />
      <div className="flex flex-wrap gap-3">
        <Link href="/" className="btn btn-primary">Home</Link>
        <Link href="/docs" className="btn btn-secondary">Documentation</Link>
      </div>
    </div>
  );
}
