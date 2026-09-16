import {
  FileText,
  Image,
  Link,
  MessageCircleQuestion,
  Newspaper,
  Search,
  Upload,
  Video,
} from "lucide-react";

import type {
  ContentTypeSummary,
  DashboardOverviewItem,
  QuickAction,
  RecentContentItem,
  SpaceSummary,
} from "@/types/dashboard";

export const dashboardOverview: DashboardOverviewItem[] = [
  {
    label: "Saved items",
    value: "24",
    detail: "Videos, documents, articles, and images",
    icon: FileText,
  },
  {
    label: "Spaces",
    value: "5",
    detail: "Topic collections ready for browsing",
    icon: Link,
  },
  {
    label: "Questions asked",
    value: "8",
    detail: "Reserved for Phase 7 RAG history",
    icon: MessageCircleQuestion,
  },
];

export const quickActions: QuickAction[] = [
  {
    title: "Save content",
    description: "Link saving arrives with ingestion.",
    icon: Link,
    disabled: true,
  },
  {
    title: "Upload file",
    description: "File upload arrives in a later phase.",
    icon: Upload,
    disabled: true,
  },
  {
    title: "Search memory",
    description: "Open the prepared search workspace.",
    icon: Search,
    href: "/search",
  },
];

export const recentContent: RecentContentItem[] = [
  {
    title: "Vector Databases Explained in 5 Minutes",
    type: "Video",
    source: "YouTube",
    metadata: "Timestamp focus: 01:42",
    space: "AI Learning",
    tag: "RAG",
    icon: Video,
  },
  {
    title: "RAG Architecture Notes.pdf",
    type: "Document",
    source: "PDF",
    metadata: "Page reference: 31",
    space: "System Design",
    tag: "AI",
    icon: FileText,
  },
  {
    title: "Choosing a Vector Database",
    type: "Article",
    source: "Web article",
    metadata: "Saved today",
    space: "AI Learning",
    tag: "Search",
    icon: Newspaper,
  },
  {
    title: "Kafka Consumer Group Diagram",
    type: "Image",
    source: "Image upload",
    metadata: "PNG mock asset",
    space: "System Design",
    tag: "Kafka",
    icon: Image,
  },
];

export const spaceSummaries: SpaceSummary[] = [
  { name: "AI Learning", itemCount: 9, lastUpdated: "Updated today" },
  { name: "System Design", itemCount: 6, lastUpdated: "Updated yesterday" },
  { name: "DSA", itemCount: 4, lastUpdated: "Updated this week" },
  { name: "Go", itemCount: 3, lastUpdated: "Updated this week" },
];

export const contentTypeSummaries: ContentTypeSummary[] = [
  {
    type: "Videos",
    count: 7,
    detail: "YouTube and short-form links",
    icon: Video,
  },
  {
    type: "Documents",
    count: 8,
    detail: "PDF, DOCX, PPTX, TXT, CSV, XLSX",
    icon: FileText,
  },
  {
    type: "Articles",
    count: 5,
    detail: "Saved web pages and references",
    icon: Newspaper,
  },
  {
    type: "Images",
    count: 4,
    detail: "Visual notes and diagrams",
    icon: Image,
  },
];
