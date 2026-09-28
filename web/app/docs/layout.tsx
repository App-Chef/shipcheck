import { DocsMobileNav, DocsSidebar } from "@/components/docs/docs-nav";

export default function DocsLayout({ children }: LayoutProps<"/docs">) {
  return (
    <div className="container-page">
      <DocsMobileNav />
      <div className="grid gap-12 py-10 lg:grid-cols-[13.5rem_minmax(0,1fr)] lg:gap-16 lg:py-14">
        <aside className="hidden lg:block">
          <DocsSidebar />
        </aside>
        {children}
      </div>
    </div>
  );
}
