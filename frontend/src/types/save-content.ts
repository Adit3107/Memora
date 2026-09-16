import type { LucideIcon } from "lucide-react";

export type SaveInputType = "youtube" | "article" | "reddit" | "file" | "image";

export type SaveInputOption = {
  value: SaveInputType;
  label: string;
  description: string;
  icon: LucideIcon;
};
