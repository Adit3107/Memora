"use client";

import {
  BookOpen,
  Bot,
  ChevronLeft,
  FileText,
  Home,
  Image,
  Library,
  MoreHorizontal,
  Newspaper,
  Plus,
  Search,
  Settings,
  User,
  Video,
  X,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useUser } from "@clerk/nextjs";

import { deleteSpace, listSpaces, updateSpace, type BackendSpace } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useEffect, useState } from "react";

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
          "relative hidden min-h-screen shrink-0 border-r bg-sidebar text-sidebar-foreground transition-[width] duration-200 lg:flex lg:flex-col",
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
      <div
        className={cn(
          "flex items-center border-b px-4 py-4",
          collapsed ? "justify-center" : "justify-between"
        )}
      >
        <Brand appHref collapsed={collapsed} />
        {!collapsed ? (
          <button
            aria-label="Collapse sidebar"
            className="inline-flex size-8 items-center justify-center rounded-md border bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            onClick={onToggle}
            type="button"
          >
            <ChevronLeft className="size-4" />
          </button>
        ) : (
          <button
            aria-label="Expand sidebar"
            className="absolute right-0 top-[1.15rem] translate-x-1/2 inline-flex size-6 items-center justify-center rounded-full border bg-background text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            onClick={onToggle}
            type="button"
          >
            <ChevronLeft className="size-3 rotate-180" />
          </button>
        )}
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
  const { user, isLoaded } = useUser();
  const router = useRouter();
  const [spaces, setSpaces] = useState<BackendSpace[]>([]);
  const displayName = user?.fullName || user?.firstName || "My Profile";
  const displayEmail = user?.primaryEmailAddress?.emailAddress || "Signed in";
  const initial = user?.firstName?.[0] || user?.fullName?.[0] || "U";

  async function renameSpace(space: BackendSpace) {
    if (!user?.id) return;
    const name = window.prompt("Rename space", space.name)?.trim();
    if (!name || name === space.name) return;
    try {
      await updateSpace(space.id, {
        user_id: user.id,
        name,
        description: space.description,
      });
      window.dispatchEvent(new Event("mindshelf:spaces-changed"));
    } catch {
      window.alert("Space could not be renamed.");
    }
  }

  async function removeSpace(space: BackendSpace) {
    if (!window.confirm(`Delete "${space.name}" and its saved content?`)) return;
    try {
      await deleteSpace(space.id);
      window.dispatchEvent(new Event("mindshelf:spaces-changed"));
      if (window.location.pathname === `/app/spaces/${space.id}`) {
        router.push("/app/spaces");
      }
    } catch {
      window.alert("Space could not be deleted.");
    }
  }

  useEffect(() => {
    if (!isLoaded || !user?.id) return;
    const userId = user.id;

    function loadUserSpaces() {
      void listSpaces(userId)
        .then(setSpaces)
        .catch(() => setSpaces([]));
    }

    loadUserSpaces();
    window.addEventListener("mindshelf:spaces-changed", loadUserSpaces);
    return () => window.removeEventListener("mindshelf:spaces-changed", loadUserSpaces);
  }, [isLoaded, user?.id]);

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
          {spaces.map((space) => (
            <div className="group/space flex items-center gap-1" key={space.id}>
              <div className="min-w-0 flex-1">
                <SidebarItem
                  collapsed={collapsed}
                  icon={BookOpen}
                  onNavigate={onNavigate}
                  title={space.name}
                  href={`/app/spaces/${space.id}`}
                />
              </div>
              {!collapsed ? (
                <details className="relative shrink-0">
                  <summary
                    aria-label={`Actions for ${space.name}`}
                    className="flex size-8 cursor-pointer list-none items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground [&::-webkit-details-marker]:hidden"
                  >
                    <MoreHorizontal className="size-4" />
                  </summary>
                  <div className="absolute right-0 top-9 z-30 grid min-w-32 gap-1 rounded-md border bg-popover p-1 shadow-lg">
                    <button
                      className="rounded px-2 py-1.5 text-left text-xs hover:bg-accent"
                      onClick={() => void renameSpace(space)}
                      type="button"
                    >
                      Rename
                    </button>
                    <button
                      className="rounded px-2 py-1.5 text-left text-xs text-destructive hover:bg-destructive/10"
                      onClick={() => void removeSpace(space)}
                      type="button"
                    >
                      Delete
                    </button>
                  </div>
                </details>
              ) : null}
            </div>
          ))}
        </div>
      </nav>

      <div className="border-t px-3 py-4">
        <SidebarItem
          collapsed={collapsed}
          href="/app/profile"
          icon={User}
          onNavigate={onNavigate}
          title="Profile"
        />
        <SidebarItem
          collapsed={collapsed}
          href="/app/settings"
          icon={Settings}
          onNavigate={onNavigate}
          title="Settings"
        />
        <div className="mt-3 flex items-center justify-between gap-2 rounded-md border bg-background px-3 py-2.5">
          <ThemeToggle />
          {!collapsed ? (
            <span className="text-xs text-muted-foreground">Theme</span>
          ) : null}
        </div>
        <Link
          className="group mt-3 flex items-center gap-3 rounded-lg border bg-background px-3 py-2.5 transition-all hover:border-primary/40 hover:bg-accent/40"
          href="/app/profile"
          onClick={onNavigate}
          title={collapsed ? displayName : undefined}
        >
          {user?.imageUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              alt=""
              className="size-8 shrink-0 rounded-full border border-border object-cover"
              src={user.imageUrl}
            />
          ) : (
            <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary text-sm font-semibold text-primary-foreground">
              {initial}
            </div>
          )}
          {!collapsed ? (
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-foreground group-hover:text-primary">
                {displayName}
              </p>
              <p className="truncate text-xs text-muted-foreground">{displayEmail}</p>
            </div>
          ) : null}
        </Link>
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
