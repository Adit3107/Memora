"use client";

import { accountNavigation, mainNavigation } from "@/data/navigation";

import { Brand } from "./brand";
import { NavLink } from "./nav-link";

export function Sidebar() {
  return (
    <aside className="hidden min-h-screen w-64 shrink-0 border-r bg-sidebar text-sidebar-foreground lg:flex lg:flex-col">
      <div className="border-b px-5 py-5">
        <Brand />
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4" aria-label="Main navigation">
        {mainNavigation.map((item) => (
          <NavLink item={item} key={item.href} />
        ))}
      </nav>

      <div className="border-t px-3 py-4">
        <div className="mb-3 rounded-lg border bg-background px-3 py-3">
          <p className="text-sm font-medium">Memora User</p>
          <p className="text-xs text-muted-foreground">Frontend placeholder</p>
        </div>
        <nav className="space-y-1" aria-label="Account navigation">
          {accountNavigation.map((item) => (
            <NavLink item={item} key={item.href} />
          ))}
        </nav>
      </div>
    </aside>
  );
}
