import type { MetadataRoute } from "next";
import { docLinks } from "@/lib/docs";
import { site } from "@/lib/site";

export default function sitemap(): MetadataRoute.Sitemap {
  const paths = ["/", ...docLinks.map((l) => l.href), "/contributing"];
  return paths.map((path) => ({
    url: new URL(path, site.url).toString(),
    changeFrequency: "monthly",
    priority: path === "/" ? 1 : path === "/docs" ? 0.8 : 0.6,
  }));
}
