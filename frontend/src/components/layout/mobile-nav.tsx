"use client";

import { Home, Library, PlusCircle, Search, Settings } from "lucide-react";

import { NavLink } from "./nav-link";

const mobileNavigation = [
  {
    title: "Home",
    href: "/app",
    icon: Home,
  },
  {
    title: "Search",
    href: "/app/search",
    icon: Search,
  },
  {
    title: "Library",
    href: "/app/library",
    icon: Library,
  },
  {
    title: "Add",
    href: "/app/save",
    icon: PlusCircle,
  },
  {
    title: "Settings",
    href: "/app/settings",
    icon: Settings,
  },
];

export function MobileNav() {
  return (
    <nav
      className="fixed inset-x-0 bottom-0 z-20 grid grid-cols-5 gap-1 border-t bg-background/95 px-2 py-2 backdrop-blur lg:hidden"
      aria-label="Mobile navigation"
    >
      {mobileNavigation.map((item) => (
        <NavLink compact item={item} key={item.href} />
      ))}
    </nav>
  );
}
