import type { LucideIcon } from "lucide-react";

export type QuickAction = {
  title: string;
  description: string;
  icon: LucideIcon;
  href?: string;
  disabled?: boolean;
};

export type RecentContentItem = {
  title: string;
  type: "Video" | "Document" | "Article" | "Image";
  source: string;
  metadata: string;
  space: string;
  tag: string;
  icon: LucideIcon;
};

export type DashboardOverviewItem = {
  label: string;
  value: string;
  detail: string;
  icon: LucideIcon;
};

export type SpaceSummary = {
  name: string;
  itemCount: number;
  lastUpdated: string;
};

export type ContentTypeSummary = {
  type: string;
  count: number;
  detail: string;
  icon: LucideIcon;
};
