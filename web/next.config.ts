import path from "node:path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // The web app is self-contained; don't let lockfiles elsewhere on the
  // machine change the workspace root.
  turbopack: {
    root: path.join(__dirname),
  },
  poweredByHeader: false,
};

export default nextConfig;
