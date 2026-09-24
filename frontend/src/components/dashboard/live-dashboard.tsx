"use client";

import Link from "next/link";
import { FileText, Image, Newspaper, Search, Video } from "lucide-react";
import { motion } from "framer-motion";
import { useEffect, useMemo, useState } from "react";

import { ContentCard } from "@/components/content/content-card";
import { EmptyState } from "@/components/content/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { buttonVariants } from "@/components/ui/button";
import { savedContent } from "@/data/content";
import {
  displayContentName,
  listContent,
  listSpaces,
  type BackendContent,
  type BackendSpace,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import type { SavedContentItem } from "@/types/content";

import { OverviewCard } from "./overview-card";
import { SectionHeading } from "./section-heading";

const iconForType = {
  video: Video,
  document: FileText,
  article: Newspaper,
  image: Image,
};

const railVariants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: 0.08,
    },
  },
};

const cardVariants = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0 },
};

export function LiveDashboard() {
  const [content, setContent] = useState<SavedContentItem[]>([]);
  const [spaces, setSpaces] = useState<BackendSpace[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    void loadDashboard();
  }, []);

  async function loadDashboard() {
    setIsLoading(true);
    setError("");
    try {
      const [contentRows, spaceRows] = await Promise.all([listContent(), listSpaces()]);
      setSpaces(spaceRows);
      setContent(contentRows.map((item) => toSavedContentItem(item, spaceRows)));
    } catch {
      setError("Could not load dashboard data from the Go backend.");
    } finally {
      setIsLoading(false);
    }
  }

  const overview = useMemo(
    () => [
      {
        label: "Saved items",
        value: String(content.length),
        detail: "Backend content records available for retrieval",
        icon: FileText,
      },
      {
        label: "Spaces",
        value: String(spaces.length),
        detail: "Topic collections loaded from the Go API",
        icon: Search,
      },
      {
        label: "Search modes",
        value: "3",
        detail: "Semantic, keyword, and hybrid retrieval",
        icon: Search,
      },
    ],
    [content.length, spaces.length]
  );

  if (isLoading) {
    return <LoadingState title="Loading dashboard" description="Fetching live Memora data." />;
  }

  if (error) {
    return (
      <ErrorState
        description={error}
        onRetry={() => void loadDashboard()}
        title="Dashboard data unavailable"
      />
    );
  }

  const recentContent = content.slice(0, 4);

  return (
    <div className="space-y-8">
      <section className="grid gap-4 md:grid-cols-3" aria-label="Overview">
        {overview.map((item) => (
          <OverviewCard item={item} key={item.label} />
        ))}
      </section>

      <section className="space-y-4" aria-labelledby="recent-content-heading">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <SectionHeading
            id="recent-content-heading"
            title="Recently added"
            description="Live content from the backend library."
          />
          <Link
            className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
            href="/library"
          >
            View library
          </Link>
        </div>

        {recentContent.length > 0 ? (
          <motion.div
            className="grid auto-cols-[minmax(280px,360px)] grid-flow-col gap-4 overflow-x-auto pb-2"
            initial="hidden"
            animate="show"
            variants={railVariants}
          >
            {recentContent.map((item) => (
              <motion.div variants={cardVariants} key={item.id ?? item.slug}>
                <ContentCard item={item} />
              </motion.div>
            ))}
          </motion.div>
        ) : (
          <EmptyState
            title="No content saved yet"
            description="Add a YouTube URL or upload a document to make this dashboard useful for testing."
          />
        )}
      </section>
    </div>
  );
}

function toSavedContentItem(
  item: BackendContent,
  spaces: BackendSpace[]
): SavedContentItem {
  const space = spaces.find((candidate) => candidate.id === item.space_id);
  const template =
    savedContent.find((mock) => mock.type === item.type)?.detail ??
    savedContent[0].detail;

  const platform = detectPlatform(item.source_url);

  return {
    id: item.id,
    slug: item.id,
    title: displayContentName(item),
    type: item.type,
    source: sourceLabel(item),
    sourceUrl: item.source_url,
    thumbnailUrl: item.thumbnail_url,
    description: item.description || "Saved to Memora and ready for search.",
    metadata: item.type === "video" ? (platform ? `${platformLabel(platform)} chunks` : "Transcript chunks") : "Extracted chunks",
    dateLabel: formatDate(item.created_at),
    spaceSlug: item.space_id,
    spaceName: space?.name ?? "Unknown space",
    tags: [],
    status: "Ready",
    icon: iconForType[item.type],
    platform,
    detail: template,
  };
}

function detectPlatform(url?: string): "instagram" | "facebook" | "youtube" | undefined {
  if (!url) return undefined;
  const lower = url.toLowerCase();
  if (lower.includes("instagram.com")) return "instagram";
  if (lower.includes("facebook.com") || lower.includes("fb.watch")) return "facebook";
  if (lower.includes("youtube.com") || lower.includes("youtu.be")) return "youtube";
  return undefined;
}

function platformLabel(platform?: string) {
  if (platform === "instagram") return "Instagram Reel";
  if (platform === "facebook") return "Facebook Reel";
  if (platform === "youtube") return "YouTube";
  return "Video";
}

function sourceLabel(item: BackendContent) {
  if (item.type === "video") {
    const platform = detectPlatform(item.source_url);
    if (platform === "instagram") return "Instagram Reel";
    if (platform === "facebook") return "Facebook Reel";
    return "YouTube";
  }
  if (item.type === "document") {
    return "Document";
  }
  if (item.type === "image") {
    return "Image";
  }
  return "Article";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}
