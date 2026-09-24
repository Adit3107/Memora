"use client";

import { FileText, Image, Newspaper, Video } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import { savedContent } from "@/data/content";
import {
  displayContentName,
  listContent,
  listContentTags,
  listSpaces,
  type BackendContent,
  type BackendSpace,
  type BackendTag,
} from "@/lib/api";
import type { ContentType, SavedContentItem } from "@/types/content";

import { ContentCard } from "./content-card";
import { EmptyState } from "./empty-state";
import { ErrorState } from "../feedback/error-state";
import { LoadingState } from "../feedback/loading-state";

type FilterValue = "all" | "youtube" | "pdf" | "reddit" | ContentType;
type SortValue = "recent" | "title" | "space";

const filters: { label: string; value: FilterValue }[] = [
  { label: "All", value: "all" },
  { label: "YouTube", value: "youtube" },
  { label: "Articles", value: "article" },
  { label: "Documents", value: "document" },
  { label: "PDFs", value: "pdf" },
  { label: "Images", value: "image" },
  { label: "Reddit", value: "reddit" },
];

const iconForType = {
  video: Video,
  document: FileText,
  article: Newspaper,
  image: Image,
};

export function LibraryBrowser() {
  const [activeFilter, setActiveFilter] = useState<FilterValue>("all");
  const [sort, setSort] = useState<SortValue>("recent");
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<SavedContentItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    void loadLibrary();
  }, []);

  async function loadLibrary() {
    setIsLoading(true);
    setError("");
    try {
      const [content, spaces] = await Promise.all([listContent(), listSpaces()]);
      const tagEntries = await Promise.all(
        content.map(async (item) => {
          try {
            return [item.id, await listContentTags(item.id)] as const;
          } catch {
            return [item.id, [] as BackendTag[]] as const;
          }
        })
      );
      const tagsByContent = new Map<string, BackendTag[]>(tagEntries);
      setItems(
        content.map((item) =>
          toSavedContentItem(item, spaces, tagsByContent.get(item.id) ?? [])
        )
      );
    } catch {
      setError("Could not load your content. Check that the Go backend is running.");
    } finally {
      setIsLoading(false);
    }
  }

  const visibleItems = useMemo(() => {
    const sourceItems = items.length > 0 ? items : [];
    const trimmedQuery = query.trim().toLowerCase();
    const filtered = sourceItems.filter((item) => {
      const matchesFilter = filterMatches(item, activeFilter);
      const matchesQuery =
        !trimmedQuery ||
        [item.title, item.description, item.source, item.spaceName, ...item.tags]
          .join(" ")
          .toLowerCase()
          .includes(trimmedQuery);

      return matchesFilter && matchesQuery;
    });

    return sortItems(filtered, sort);
  }, [activeFilter, items, query, sort]);

  if (isLoading) {
    return (
      <LoadingState
        title="Loading library"
        description="Fetching content, spaces, and tags from the Go API."
      />
    );
  }

  if (error) {
    return (
      <ErrorState
        description={error}
        onRetry={() => void loadLibrary()}
        title="Library could not be loaded"
      />
    );
  }

  return (
    <div className="space-y-5">
      <section className="space-y-4 rounded-md border bg-card p-4 shadow-sm">
        <input
          className="h-11 w-full rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search this library..."
          value={query}
        />
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div
            className="flex gap-2 overflow-x-auto pb-1"
            aria-label="Content filters"
          >
            {filters.map((filter) => {
              const isActive = activeFilter === filter.value;

              return (
                <button
                  className={[
                    "rounded-md border px-3 py-2 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-primary text-primary-foreground"
                      : "bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                  ].join(" ")}
                  key={filter.value}
                  onClick={() => setActiveFilter(filter.value)}
                  type="button"
                >
                  {filter.label}
                </button>
              );
            })}
          </div>

          <label className="flex w-full flex-col gap-2 text-sm font-medium sm:w-56">
            Sort
            <select
              className="h-9 rounded-md border bg-background px-3 text-sm text-foreground outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setSort(event.target.value as SortValue)}
              value={sort}
            >
              <option value="recent">Recently saved</option>
              <option value="title">Title</option>
              <option value="space">Space</option>
            </select>
          </label>
        </div>
      </section>

      <div className="flex items-center justify-between gap-4">
        <p className="text-sm text-muted-foreground">
          Showing {visibleItems.length}{" "}
          {activeFilter === "all" ? "items" : "items"}
        </p>
      </div>

      {visibleItems.length > 0 ? (
        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {visibleItems.map((item) => (
            <ContentCard item={item} key={item.id ?? item.slug} />
          ))}
        </section>
      ) : (
        <EmptyState
          title="No memories yet"
          description="Save a YouTube URL or upload a document to test ingestion and search."
        />
      )}
    </div>
  );
}

function filterMatches(item: SavedContentItem, filter: FilterValue) {
  if (filter === "all") {
    return true;
  }
  if (filter === "youtube") {
    return item.type === "video" || item.source.toLowerCase().includes("youtube");
  }
  if (filter === "pdf") {
    return item.type === "document" && item.title.toLowerCase().endsWith(".pdf");
  }
  if (filter === "reddit") {
    return item.source.toLowerCase().includes("reddit");
  }
  return item.type === filter;
}

function sortItems(items: SavedContentItem[], sort: SortValue) {
  return [...items].sort((a, b) => {
    if (sort === "title") {
      return a.title.localeCompare(b.title);
    }

    if (sort === "space") {
      return a.spaceName.localeCompare(b.spaceName);
    }

    return 0;
  });
}

function toSavedContentItem(
  item: BackendContent,
  spaces: BackendSpace[],
  tags: BackendTag[]
): SavedContentItem {
  const space = spaces.find((candidate) => candidate.id === item.space_id);
  const type = item.type;
  const displayName = displayContentName(item);

  const platform = detectPlatform(item.source_url);

  return {
    id: item.id,
    slug: item.id,
    title: displayName,
    type,
    source: sourceLabel(item),
    sourceUrl: item.source_url,
    thumbnailUrl: item.thumbnail_url,
    description: item.description || "Saved to Memora and ready for search.",
    metadata: metadataLabel(item),
    dateLabel: formatDate(item.created_at),
    spaceSlug: item.space_id,
    spaceName: space?.name ?? "Unknown space",
    tags: tags.map((tag) => tag.name),
    status: "Ready",
    icon: iconForType[type],
    platform,
    detail: savedContent.find((mock) => mock.type === type)?.detail ?? savedContent[0].detail,
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

function metadataLabel(item: BackendContent) {
  if (item.type === "video") {
    const platform = detectPlatform(item.source_url);
    if (platform === "instagram") return "Instagram Reel chunks";
    if (platform === "facebook") return "Facebook Reel chunks";
    return "Transcript chunks";
  }
  if (item.type === "document") {
    return "Extracted document text";
  }
  return "Extracted source text";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}
