"use client";

import {
  BookOpen,
  Bot,
  ChevronLeft,
  FileText,
  Home,
  Image,
  Library,
  Newspaper,
  Plus,
  Search,
  Settings,
  Video,
  X,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { spaces } from "@/data/content";
import { cn } from "@/lib/utils";

import { ThemeToggle } from "../theme/theme-toggle";
import { Brand } from "./brand";

type SidebarProps = {
  collapsed?: boolean;
  mobileOpen?: boolean;
  onCloseMobile?: () => void;
  onToggle?: () => void;
};

const primaryNavigation = [
  { title: "Home", href: "/app", icon: Home },
  { title: "AI Playground", href: "/app/ai", icon: Bot, badge: "Soon" },
  { title: "Search", href: "/app/search", icon: Search },
  { title: "Library", href: "/app/library", icon: Library },
];

const contentNavigation = [
  { title: "Videos", href: "/app/videos", icon: Video },
  { title: "Articles", href: "/app/articles", icon: Newspaper },
  { title: "Documents", href: "/app/documents", icon: FileText },
  { title: "Images", href: "/app/images", icon: Image },
];

export function Sidebar({
  collapsed = false,
  mobileOpen = false,
  onCloseMobile,
  onToggle,
}: SidebarProps) {
  return (
    <>
      <aside
        className={cn(
          "hidden min-h-screen shrink-0 border-r bg-sidebar text-sidebar-foreground transition-[width] duration-200 lg:flex lg:flex-col",
          collapsed ? "w-20" : "w-72"
        )}
      >
        <SidebarContent collapsed={collapsed} onToggle={onToggle} />
      </aside>

      {mobileOpen ? (
        <div className="fixed inset-0 z-40 lg:hidden">
          <button
            aria-label="Close navigation"
            className="absolute inset-0 bg-foreground/20"
            onClick={onCloseMobile}
            type="button"
          />
          <aside className="absolute inset-y-0 left-0 flex w-80 max-w-[85vw] flex-col border-r bg-sidebar text-sidebar-foreground shadow-xl">
            <div className="flex items-center justify-between border-b px-5 py-4">
              <Brand appHref />
              <button
                aria-label="Close navigation"
                className="inline-flex size-8 items-center justify-center rounded-md border bg-background"
                onClick={onCloseMobile}
                type="button"
              >
                <X className="size-4" />
              </button>
            </div>
            <SidebarNav collapsed={false} onNavigate={onCloseMobile} />
          </aside>
        </div>
      ) : null}
    </>
  );
}

function SidebarContent({
  collapsed,
  onToggle,
}: {
  collapsed: boolean;
  onToggle?: () => void;
}) {
  return (
    <>
      <div className="flex items-center justify-between border-b px-4 py-4">
        <Brand appHref collapsed={collapsed} />
        <button
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="inline-flex size-8 items-center justify-center rounded-md border bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
          onClick={onToggle}
          type="button"
        >
          <ChevronLeft
            className={cn("size-4 transition-transform", collapsed && "rotate-180")}
          />
        </button>
      </div>
      <SidebarNav collapsed={collapsed} />
    </>
  );
}

function SidebarNav({
  collapsed,
  onNavigate,
}: {
  collapsed: boolean;
  onNavigate?: () => void;
}) {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <nav className="flex-1 space-y-6 overflow-y-auto px-3 py-4" aria-label="Main navigation">
        <Link
          className={cn(
            "flex items-center justify-center gap-2 rounded-xl bg-primary px-3 py-3 text-sm font-semibold text-primary-foreground shadow-sm transition-transform hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50",
            collapsed && "px-2"
          )}
          href="/app/save"
          onClick={onNavigate}
          title={collapsed ? "Add Content" : undefined}
        >
          <Plus className="size-4" />
          {!collapsed ? <span>Add Content</span> : null}
        </Link>
        <NavGroup collapsed={collapsed} items={primaryNavigation} onNavigate={onNavigate} />
        <NavGroup
          collapsed={collapsed}
          items={contentNavigation}
          label="Content"
          onNavigate={onNavigate}
        />
        <div className="space-y-2">
          <GroupLabel collapsed={collapsed}>Spaces</GroupLabel>
          <SidebarItem
            collapsed={collapsed}
            icon={Plus}
            onNavigate={onNavigate}
            title="Create Space"
            href="/app/spaces"
          />
          {spaces.slice(0, 4).map((space) => (
            <SidebarItem
              collapsed={collapsed}
              icon={BookOpen}
              key={space.slug}
              onNavigate={onNavigate}
              title={space.name}
              href={`/app/spaces/${space.slug}`}
            />
          ))}
        </div>
      </nav>

      <div className="border-t px-3 py-4">
        <SidebarItem
          collapsed={collapsed}
          href="/app/settings"
          icon={Settings}
          onNavigate={onNavigate}
          title="Settings"
        />
        <div className="mt-3 flex items-center justify-between gap-2 rounded-md border bg-background px-3 py-3">
          <ThemeToggle />
          {!collapsed ? (
            <span className="text-xs text-muted-foreground">Theme</span>
          ) : null}
        </div>
        <div className="mt-3 flex items-center gap-3 rounded-md border bg-background px-3 py-3">
          <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
            A
          </div>
          {!collapsed ? (
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">Aditya</p>
              <p className="truncate text-xs text-muted-foreground">Local demo</p>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}

function NavGroup({
  collapsed,
  items,
  label,
  onNavigate,
}: {
  collapsed: boolean;
  items: typeof primaryNavigation;
  label?: string;
  onNavigate?: () => void;
}) {
  return (
    <div className="space-y-2">
      {label ? <GroupLabel collapsed={collapsed}>{label}</GroupLabel> : null}
      {items.map((item) => (
        <SidebarItem
          badge={"badge" in item ? (item.badge as string) : undefined}
          collapsed={collapsed}
          href={item.href}
          icon={item.icon}
          key={item.href}
          onNavigate={onNavigate}
          title={item.title}
        />
      ))}
    </div>
  );
}

function GroupLabel({
  children,
  collapsed,
}: {
  children: string;
  collapsed: boolean;
}) {
  if (collapsed) {
    return <div className="h-px bg-border" />;
  }

  return (
    <p className="px-3 text-xs font-semibold uppercase tracking-normal text-muted-foreground">
      {children}
    </p>
  );
}

function SidebarItem({
  badge,
  collapsed,
  href,
  icon: Icon,
  onNavigate,
  title,
}: {
  badge?: string;
  collapsed: boolean;
  href: string;
  icon: typeof Home;
  onNavigate?: () => void;
  title: string;
}) {
  const pathname = usePathname();
  const isActive = pathname === href || pathname.startsWith(`${href}/`);

  return (
    <Link
      aria-current={isActive ? "page" : undefined}
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50",
        isActive && "bg-accent text-accent-foreground",
        collapsed && "justify-center px-2"
      )}
      href={href}
      onClick={onNavigate}
      title={collapsed ? title : undefined}
    >
      <Icon className="size-4 shrink-0" aria-hidden="true" />
      {!collapsed ? (
        <>
          <span className="truncate">{title}</span>
          {badge ? (
            <span className="ml-auto rounded-full border border-primary/30 bg-primary/10 px-1.5 py-0.5 text-[10px] font-semibold text-primary">
              {badge}
            </span>
          ) : null}
        </>
      ) : null}
    </Link>
  );
}
