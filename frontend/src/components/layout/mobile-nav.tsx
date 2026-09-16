"use client";

import { accountNavigation, mainNavigation } from "@/data/navigation";

import { Brand } from "./brand";
import { NavLink } from "./nav-link";

export function MobileNav() {
  return (
    <>
      <header className="sticky top-0 z-20 border-b bg-background/95 px-4 py-3 backdrop-blur lg:hidden">
        <div className="flex items-center justify-between gap-4">
          <Brand />
          <div className="rounded-md border px-3 py-2 text-xs font-medium text-muted-foreground">
            Phase 1
          </div>
        </div>
      </header>

      <nav
        className="fixed inset-x-0 bottom-0 z-20 grid grid-cols-[repeat(7,minmax(0,1fr))] gap-1 border-t bg-background/95 px-2 py-2 backdrop-blur lg:hidden"
        aria-label="Mobile navigation"
      >
        {[...mainNavigation, ...accountNavigation].map((item) => (
          <NavLink compact item={item} key={item.href} />
        ))}
      </nav>
    </>
  );
}
