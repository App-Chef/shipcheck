export type DocLink = {
  href: string;
  title: string;
  description: string;
};

export type DocGroup = {
  title: string;
  links: DocLink[];
};

export const docGroups: DocGroup[] = [
  {
    title: "Getting started",
    links: [
      {
        href: "/docs",
        title: "Introduction",
        description: "What Shipcheck is and what it checks.",
      },
      {
        href: "/docs/installation",
        title: "Installation",
        description: "Install the CLI with Go or build it from source.",
      },
      {
        href: "/docs/quick-start",
        title: "Quick start",
        description: "Run your first check and read the results.",
      },
    ],
  },
  {
    title: "Reference",
    links: [
      {
        href: "/docs/configuration",
        title: "Configuration",
        description: "Every option in .shipcheck.yml.",
      },
      {
        href: "/docs/checks",
        title: "Checks",
        description: "What each check looks for and why.",
      },
      {
        href: "/docs/ci",
        title: "CI",
        description: "Run Shipcheck in GitHub Actions and GitLab CI.",
      },
      {
        href: "/docs/json-output",
        title: "JSON output",
        description: "The machine-readable report format.",
      },
      {
        href: "/docs/exit-codes",
        title: "Exit codes",
        description: "What 0, 1 and 2 mean.",
      },
    ],
  },
  {
    title: "Project",
    links: [
      {
        href: "/docs/security",
        title: "Security model",
        description: "What Shipcheck runs, and what it never does.",
      },
    ],
  },
];

export const docLinks: DocLink[] = docGroups.flatMap((g) => g.links);

export function docNeighbours(href: string) {
  const i = docLinks.findIndex((l) => l.href === href);
  return {
    prev: i > 0 ? docLinks[i - 1] : undefined,
    next: i >= 0 && i < docLinks.length - 1 ? docLinks[i + 1] : undefined,
  };
}
