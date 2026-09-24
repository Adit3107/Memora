"use client";

import { useUser } from "@clerk/nextjs";
import { BookOpen, FileText, Image, Newspaper, Video } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  displayContentName,
  listContent,
  listContentTags,
  listSpaces,
  type BackendContent,
  type BackendSpace,
  type BackendTag,
} from "@/lib/api";
import { savedContent } from "@/data/content";
import type { SavedContentItem } from "@/types/content";

import { ContentCard } from "./content-card";
import { EmptyState } from "./empty-state";
import { ErrorState } from "../feedback/error-state";
import { LoadingState } from "../feedback/loading-state";

type FilterValue = "all" | "video" | "document" | "article";
type SortValue = "recent" | "title" | "space";
type SourceValue = "all" | "youtube" | "instagram" | "facebook" | string;

const filters: { label: string; value: FilterValue }[] = [
  { label: "All", value: "all" },
  { label: "Videos", value: "video" },
  { label: "Articles", value: "article" },
  { label: "Documents", value: "document" },
];

const iconForType = {
  video: Video,
  document: FileText,
  article: Newspaper,
  image: Image,
};

export function LibraryBrowser() {
  const { user, isLoaded } = useUser();
  const [activeFilter, setActiveFilter] = useState<FilterValue>("all");
  const [sort, setSort] = useState<SortValue>("recent");
  const [sourceFilter, setSourceFilter] = useState<SourceValue>("all");
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<SavedContentItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isLoaded) void loadLibrary();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoaded, user?.id]);

  async function loadLibrary() {
    setIsLoading(true);
    setError("");
    try {
      const userId = user?.id;
      const [content, spaces] = await Promise.all([
        listContent(userId),
        listSpaces(userId),
      ]);
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
      const matchesSource = sourceMatches(item, sourceFilter);
      const matchesQuery =
        !trimmedQuery ||
        [item.title, item.description, item.source, item.spaceName, ...item.tags]
          .join(" ")
          .toLowerCase()
          .includes(trimmedQuery);

      return matchesFilter && matchesSource && matchesQuery;
    });

    return sortItems(filtered, sort);
  }, [activeFilter, items, query, sort, sourceFilter]);

  const sourceOptions = sourceOptionsFor(activeFilter, items);

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
                  onClick={() => {
                    setActiveFilter(filter.value);
                    setSourceFilter("all");
                  }}
                  type="button"
                >
                  {filter.label}
                </button>
              );
            })}
          </div>

          <div className="grid w-full gap-3 sm:w-auto sm:grid-cols-2">
            <label className="flex min-w-40 flex-col gap-2 text-sm font-medium">
              Source
              <select
                className="h-9 rounded-md border bg-background px-3 text-sm text-foreground outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                onChange={(event) => setSourceFilter(event.target.value)}
                value={sourceFilter}
              >
                {sourceOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex min-w-40 flex-col gap-2 text-sm font-medium">
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
            <ContentCard
              item={item}
              key={item.id ?? item.slug}
              onDeleted={(contentID) =>
                setItems((current) => current.filter((entry) => entry.id !== contentID))
              }
            />
          ))}
        </section>
      ) : (
        <div className="flex flex-col items-center justify-center gap-6 rounded-2xl border border-dashed border-border bg-card/50 py-20 text-center">
          <div className="flex h-20 w-20 items-center justify-center rounded-full bg-primary/10">
            <BookOpen className="h-10 w-10 text-primary" />
          </div>
          <div className="space-y-2">
            <h3 className="text-xl font-semibold text-foreground">No saved content yet</h3>
            <p className="max-w-sm text-sm text-muted-foreground">
              Start building your second brain — save a YouTube video, paste an article URL, or upload a document.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

function filterMatches(item: SavedContentItem, filter: FilterValue) {
  if (filter === "all") {
    return true;
  }
  return item.type === filter;
}

function sourceMatches(item: SavedContentItem, source: SourceValue) {
  if (source === "all") return true;
  if (item.type === "document") {
    return documentExtension(item.title).toLowerCase() === source;
  }
  return item.platform === source;
}

function sourceOptionsFor(filter: FilterValue, items: SavedContentItem[]) {
  if (filter === "video") {
    return [
      { label: "All video sources", value: "all" },
      { label: "YouTube", value: "youtube" },
      { label: "Instagram", value: "instagram" },
      { label: "Facebook", value: "facebook" },
    ];
  }

  if (filter === "document") {
    const extensions = Array.from(
      new Set(
        items
          .filter((item) => item.type === "document")
          .map((item) => documentExtension(item.title))
      )
    );

    return [
      { label: "All document types", value: "all" },
      ...extensions.map((extension) => ({
        label: extension,
        value: extension.toLowerCase(),
      })),
    ];
  }

  return [{ label: "All sources", value: "all" }];
}

function documentExtension(title: string) {
  const extension = title.split(".").pop();
  return extension && extension.length <= 5 ? extension.toUpperCase() : "DOC";
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
    description: item.description || "Saved to Mindshelf and ready for search.",
    metadata: metadataLabel(item),
    dateLabel: formatDate(item.created_at),
    spaceSlug: item.space_id,
    spaceName: space?.name ?? "Unknown space",
    tags: tags.map((tag) => tag.name),
    status: "Ready",
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
