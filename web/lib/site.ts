export const site = {
  name: "Shipcheck",
  title: "Shipcheck | Know you're ready before you ship",
  description:
    "A minimal open-source CLI that checks your project for common release blockers.",
  // Set NEXT_PUBLIC_SITE_URL to the deployed URL. The fallback is only for
  // local development and must be replaced once the site is live.
  url: process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000",
  repo: "https://github.com/App-Chef/shipcheck",
  license: "MIT",
  minGo: "1.22",
  installCommand:
    "go install github.com/App-Chef/shipcheck/cli/cmd/shipcheck@latest",
} as const;

export const repoFile = (path: string) => `${site.repo}/blob/main/${path}`;
