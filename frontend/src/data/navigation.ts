import {
  BookOpen,
  Home,
  Library,
  PlusCircle,
  Search,
  Settings,
  UserRound,
} from "lucide-react";

import type { NavigationItem } from "@/types/navigation";

export const mainNavigation: NavigationItem[] = [
  {
    title: "Dashboard",
    href: "/",
    icon: Home,
    description: "Overview of saved knowledge",
  },
  {
    title: "Spaces",
    href: "/spaces",
    icon: BookOpen,
    description: "Topic-based collections",
  },
  {
    title: "Library",
    href: "/library",
    icon: Library,
    description: "All saved content",
  },
  {
    title: "Save",
    href: "/save",
    icon: PlusCircle,
    description: "Save a link or upload",
  },
  {
    title: "Search",
    href: "/search",
    icon: Search,
    description: "Find saved knowledge",
  },
];

export const accountNavigation: NavigationItem[] = [
  {
    title: "Profile",
    href: "/profile",
    icon: UserRound,
  },
  {
    title: "Settings",
    href: "/settings",
    icon: Settings,
  },
];
