"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { Close, GitHub, Menu } from "./icons";
import { LogoMark } from "./logo";
import { site } from "@/lib/site";

const nav = [
  { href: "/docs", label: "Docs", match: (p: string) => p.startsWith("/docs") && p !== "/docs/checks" },
  { href: "/docs/checks", label: "Checks", match: (p: string) => p === "/docs/checks" },
];

export function SiteHeader() {
  const pathname = usePathname() ?? "/";
  const [open, setOpen] = useState(false);
  const toggleRef = useRef<HTMLButtonElement>(null);

  // Close the menu when navigating.
  const [lastPath, setLastPath] = useState(pathname);
  if (pathname !== lastPath) {
    setLastPath(pathname);
    setOpen(false);
  }

  useEffect(() => {
    if (!open) return;
    function onKey(e: globalThis.KeyboardEvent) {
      if (e.key === "Escape") {
        setOpen(false);
        toggleRef.current?.focus();
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [open]);

  return (
    <header className="sticky top-0 z-40 border-b border-ink bg-paper/90 backdrop-blur-sm supports-[backdrop-filter]:bg-paper/80">
      <div className="container-page flex h-16 items-center justify-between gap-6">
        <Link
          href="/"
          className="flex items-center gap-2.5 rounded-sm text-[17px] font-semibold tracking-tight"
          aria-label="Shipcheck home"
        >
          <LogoMark />
          Shipcheck
        </Link>

        <nav aria-label="Main" className="hidden items-center gap-1 md:flex">
          {nav.map((item) => {
            const active = item.match(pathname);
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={active ? "page" : undefined}
                className={`relative rounded-sm px-3 py-2 text-[15px] transition-colors duration-150 ease-out ${
                  active ? "text-ink" : "text-muted hover:text-ink"
                }`}
              >
                {item.label}
                <span
                  aria-hidden
                  className={`absolute inset-x-3 -bottom-[13px] h-[2px] bg-ink transition-opacity duration-200 ease-out ${
                    active ? "opacity-100" : "opacity-0"
                  }`}
                />
              </Link>
            );
          })}
          <a
            href={site.repo}
            className="flex items-center gap-2 rounded-sm px-3 py-2 text-[15px] text-muted transition-colors duration-150 ease-out hover:text-ink"
          >
            <GitHub />
            GitHub
          </a>
          <Link href="/docs/installation" className="btn btn-primary btn-sm ml-3">
            Install
          </Link>
        </nav>

        <button
          ref={toggleRef}
          type="button"
          className="-mr-2 grid size-10 place-items-center rounded-md md:hidden"
          aria-expanded={open}
          aria-controls="mobile-menu"
          aria-label={open ? "Close menu" : "Open menu"}
          onClick={() => setOpen((o) => !o)}
        >
          {open ? <Close width={20} height={20} /> : <Menu width={20} height={20} />}
        </button>
      </div>

      <div
        id="mobile-menu"
        hidden={!open}
        className="border-t border-subtle bg-paper md:hidden"
      >
        <nav aria-label="Mobile" className="container-page flex flex-col py-3 motion-safe:animate-[line-in_200ms_ease-out]">
          {nav.map((item) => {
            const active = item.match(pathname);
            return (
              <Link
                key={item.href}
                href={item.href}
                aria-current={active ? "page" : undefined}
                className={`flex items-center justify-between border-b border-subtle py-3.5 text-lg ${
                  active ? "font-semibold" : ""
                }`}
              >
                {item.label}
                {active && <span className="size-2 rounded-full bg-ink" aria-hidden />}
              </Link>
            );
          })}
          <a href={site.repo} className="flex items-center gap-2 border-b border-subtle py-3.5 text-lg">
            <GitHub width={18} height={18} />
            GitHub
          </a>
          <Link href="/docs/installation" className="btn btn-primary mt-5 mb-2 w-full">
            Install Shipcheck
          </Link>
        </nav>
      </div>
    </header>
  );
}
