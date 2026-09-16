"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";
import type { NavigationItem } from "@/types/navigation";

type NavLinkProps = {
  item: NavigationItem;
  compact?: boolean;
};

export function NavLink({ item, compact = false }: NavLinkProps) {
  const pathname = usePathname();
  const isActive =
    item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
  const Icon = item.icon;

  return (
    <Link
      aria-current={isActive ? "page" : undefined}
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none",
        isActive && "bg-accent text-accent-foreground",
        compact && "justify-center gap-1 px-2 py-2 text-xs"
      )}
      href={item.href}
    >
      <Icon className="size-4 shrink-0" aria-hidden="true" />
      <span className={compact ? "sr-only sm:not-sr-only" : undefined}>
        {item.title}
      </span>
    </Link>
  );
}
