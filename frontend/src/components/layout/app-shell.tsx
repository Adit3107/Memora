"use client";

import { Menu, Plus, Search } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { useUser } from "@clerk/nextjs";

import { GlobalSearchDialog } from "@/components/search/global-search-dialog";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

import { MobileNav } from "./mobile-nav";
import { Sidebar } from "./sidebar";

type AppShellProps = {
  children: ReactNode;
};

const publicRoutes = ["/", "/login", "/signup", "/forgot-password", "/sso-callback"];

export function AppShell({ children }: AppShellProps) {
  const pathname = usePathname();
  const { user } = useUser();
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMobileOpen, setIsMobileOpen] = useState(false);
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const isPublicRoute = publicRoutes.some((route) => pathname === route || pathname.startsWith(`${route}/`));

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setIsSearchOpen(true);
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  if (isPublicRoute) {
    return <div className="min-h-screen bg-background text-foreground">{children}</div>;
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="flex min-h-screen">
        <Sidebar
          collapsed={isCollapsed}
          mobileOpen={isMobileOpen}
          onCloseMobile={() => setIsMobileOpen(false)}
          onToggle={() => setIsCollapsed((current) => !current)}
        />
        <div className="flex min-w-0 flex-1 flex-col">
          <header className="sticky top-0 z-20 border-b bg-background/85 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-background/70 sm:px-6 lg:px-8">
            <div className="flex items-center gap-3">
              <button
                aria-label="Open navigation"
                className="inline-flex size-9 items-center justify-center rounded-md border bg-card lg:hidden"
                onClick={() => setIsMobileOpen(true)}
                type="button"
              >
                <Menu className="size-4" />
              </button>

              <div className="hidden min-w-24 text-sm font-semibold lg:block">
                {titleForPath(pathname)}
              </div>

              <button
                className="flex h-10 min-w-0 flex-1 items-center gap-3 rounded-md border bg-card px-3 text-left text-sm text-muted-foreground shadow-sm transition-colors hover:bg-accent/60"
                onClick={() => setIsSearchOpen(true)}
                type="button"
              >
                <Search className="size-4 shrink-0" />
                <span className="truncate">Search your memory...</span>
                <kbd className="ml-auto hidden rounded-md border bg-background px-2 py-1 font-mono text-xs sm:inline-flex">
                  Ctrl K
                </kbd>
              </button>

              <ThemeToggle />

              <Link
                className={cn(buttonVariants({ variant: "default" }), "hidden sm:inline-flex")}
                href="/app/save"
              >
                <Plus className="size-4" />
                Add memory
              </Link>

              <Link
                aria-label="Profile"
                className="flex size-9 items-center justify-center overflow-hidden rounded-full border bg-card text-sm font-semibold transition-transform hover:scale-105"
                href="/app/profile"
                title={user?.fullName || user?.firstName || "Profile"}
              >
                {user?.imageUrl ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    alt="Profile"
                    className="size-full object-cover"
                    src={user.imageUrl}
                  />
                ) : (
                  <span>{user?.firstName?.[0] || user?.fullName?.[0] || "U"}</span>
                )}
              </Link>
            </div>
          </header>

          <MobileNav />

          <main className="flex-1 px-4 py-6 pb-24 sm:px-6 lg:px-8 lg:py-8 lg:pb-8">
            {children}
          </main>
        </div>
      </div>
      <GlobalSearchDialog open={isSearchOpen} onOpenChange={setIsSearchOpen} />
    </div>
  );
}

function titleForPath(pathname: string) {
  if (pathname.includes("/profile")) {
    return "Profile";
  }
  if (pathname.includes("/search")) {
    return "Search";
  }
  if (pathname.includes("/ai")) {
    return "AI Playground";
  }
  if (pathname.includes("/library")) {
    return "Library";
  }
  if (pathname.includes("/spaces")) {
    return "Spaces";
  }
  if (pathname.includes("/settings")) {
    return "Settings";
  }
  if (pathname.includes("/save")) {
    return "Add memory";
  }
  return "Home";
}
