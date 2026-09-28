"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { ChevronDown } from "../icons";
import { docGroups, docLinks } from "@/lib/docs";

function NavList({ pathname, onNavigate }: { pathname: string; onNavigate?: () => void }) {
  return (
    <div className="space-y-7">
      {docGroups.map((group) => (
        <div key={group.title}>
          <h2 className="eyebrow mb-2.5 px-3">{group.title}</h2>
          <ul className="space-y-0.5">
            {group.links.map((link) => {
              const active = pathname === link.href;
              return (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    onClick={onNavigate}
                    aria-current={active ? "page" : undefined}
                    className={`block rounded-[4px] border px-3 py-1.5 text-[15px] transition-[background-color,color,border-color] duration-150 ease-out ${
                      active
                        ? "border-ink bg-card font-medium text-ink shadow-hard-sm"
                        : "border-transparent text-muted hover:bg-wash hover:text-ink"
                    }`}
                  >
                    {link.title}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      ))}
    </div>
  );
}

/** Sticky sidebar for large screens. */
export function DocsSidebar() {
  const pathname = usePathname() ?? "/docs";
  return (
    <nav aria-label="Documentation" className="sticky top-24 max-h-[calc(100dvh-7rem)] overflow-y-auto pb-8">
      <NavList pathname={pathname} />
    </nav>
  );
}

/** Disclosure control that replaces the sidebar below the lg breakpoint. */
export function DocsMobileNav() {
  const pathname = usePathname() ?? "/docs";
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const current = docLinks.find((l) => l.href === pathname);

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
        buttonRef.current?.focus();
      }
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [open]);

  return (
    <div className="sticky top-16 z-30 -mx-5 border-b border-ink bg-paper/95 px-5 py-3 backdrop-blur-sm sm:-mx-8 sm:px-8 lg:hidden">
      <button
        ref={buttonRef}
        type="button"
        aria-expanded={open}
        aria-controls="docs-mobile-nav"
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between gap-3 rounded-md border border-ink bg-card px-3.5 py-2.5 text-left shadow-hard-sm"
      >
        <span className="min-w-0">
          <span className="eyebrow block text-[10px]">Documentation</span>
          <span className="block truncate font-medium">{current?.title ?? "Menu"}</span>
        </span>
        <ChevronDown
          width={18}
          height={18}
          className={`shrink-0 transition-transform duration-200 ease-out ${open ? "rotate-180" : ""}`}
        />
      </button>
      <nav
        id="docs-mobile-nav"
        aria-label="Documentation"
        hidden={!open}
        className="mt-3 max-h-[65dvh] overflow-y-auto rounded-md border border-ink bg-card p-3 shadow-hard motion-safe:animate-[line-in_180ms_ease-out]"
      >
        <NavList pathname={pathname} onNavigate={() => setOpen(false)} />
      </nav>
    </div>
  );
}
