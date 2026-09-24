"use client";

import { Search } from "lucide-react";
import { FormEvent, useEffect, useMemo, useState } from "react";

import { EmptyState } from "@/components/content/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import {
  listSpaces,
  listTags,
  MEMORA_DEMO_USER_ID,
  searchMemora,
  type BackendSpace,
  type BackendTag,
  type SearchMode,
  type SearchResult,
} from "@/lib/api";
import type { ContentType } from "@/types/content";

import { SearchResultCard } from "./search-result-card";

const contentTypeOptions: { label: string; value: "all" | ContentType }[] = [
  { label: "All", value: "all" },
  { label: "Videos", value: "video" },
  { label: "Documents", value: "document" },
  { label: "Articles", value: "article" },
  { label: "Images", value: "image" },
];

const modeOptions: { label: string; value: SearchMode }[] = [
  { label: "Hybrid", value: "hybrid" },
  { label: "Semantic", value: "semantic" },
  { label: "Keyword", value: "keyword" },
];

const sourceTypeOptions = [
  { label: "All sources", value: "all" },
  { label: "YouTube", value: "youtube" },
  { label: "Video chunks", value: "video" },
  { label: "Documents", value: "document" },
  { label: "Text files", value: "text" },
];

export function SearchBrowser() {
  const [query, setQuery] = useState("");
  const [spaces, setSpaces] = useState<BackendSpace[]>([]);
  const [tags, setTags] = useState<BackendTag[]>([]);
  const [spaceID, setSpaceID] = useState("all");
  const [tagID, setTagID] = useState("all");
  const [mode, setMode] = useState<SearchMode>("hybrid");
  const [contentType, setContentType] = useState<"all" | ContentType>("all");
  const [sourceType, setSourceType] = useState("all");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [total, setTotal] = useState(0);
  const [hasSearched, setHasSearched] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState("");

  async function loadFilters() {
    try {
      const [spaceRows, tagRows] = await Promise.all([listSpaces(), listTags()]);
      setSpaces(spaceRows);
      setTags(tagRows);
    } catch {
      setError("Search filters could not be loaded. Search still works if the backend is available.");
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadFilters();
  }, []);

  const visibleResults = useMemo(() => dedupeByContent(results), [results]);

  async function runSearch(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();
    const trimmedQuery = query.trim();
    if (!trimmedQuery) {
      setResults([]);
      setTotal(0);
      setHasSearched(false);
      setError("");
      return;
    }

    setIsSearching(true);
    setError("");
    try {
      const response = await searchMemora({
        user_id: MEMORA_DEMO_USER_ID,
        space_id: spaceID === "all" ? undefined : spaceID,
        query: trimmedQuery,
        mode,
        content_type: contentType === "all" ? undefined : contentType,
        source_type: sourceType === "all" ? undefined : sourceType,
        tag_ids: tagID === "all" ? [] : [tagID],
        limit: 10,
        offset: 0,
      });
      setResults(response.results);
      setTotal(response.total);
      setHasSearched(true);
    } catch {
      setResults([]);
      setTotal(0);
      setHasSearched(true);
      setError("Search couldn't be completed. Check the backend, AI service, and database, then try again.");
    } finally {
      setIsSearching(false);
    }
  }

  return (
    <div className="space-y-6">
      <form className="space-y-4 rounded-md border bg-card p-5 shadow-sm" onSubmit={runSearch}>
        <label className="text-sm font-medium" htmlFor="memora-search">
          Search your memory
        </label>
        <div className="mt-3 flex items-center gap-3 rounded-md border bg-background px-3 py-2">
          <Search className="size-4 text-muted-foreground" aria-hidden="true" />
          <input
            className="h-11 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            id="memora-search"
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search your saved knowledge..."
            type="search"
            value={query}
          />
        </div>

        <div className="grid gap-3 md:grid-cols-5">
          <label className="space-y-2 text-sm font-medium">
            Mode
            <select
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setMode(event.target.value as SearchMode)}
              value={mode}
            >
              {modeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-2 text-sm font-medium">
            Space
            <select
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setSpaceID(event.target.value)}
              value={spaceID}
            >
              <option value="all">All Spaces</option>
              {spaces.map((space) => (
                <option key={space.id} value={space.id}>
                  {space.name}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-2 text-sm font-medium">
            Type
            <select
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) =>
                setContentType(event.target.value as "all" | ContentType)
              }
              value={contentType}
            >
              {contentTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-2 text-sm font-medium">
            Source
            <select
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setSourceType(event.target.value)}
              value={sourceType}
            >
              {sourceTypeOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-2 text-sm font-medium">
            Tag
            <select
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setTagID(event.target.value)}
              value={tagID}
            >
              <option value="all">All Tags</option>
              {tags.map((tag) => (
                <option key={tag.id} value={tag.id}>
                  {tag.name}
                </option>
              ))}
            </select>
          </label>
        </div>

        <button
          className="h-10 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSearching || !query.trim()}
          type="submit"
        >
          {isSearching ? "Searching..." : "Search"}
        </button>
      </form>

      {isSearching ? <SearchSkeleton /> : null}

      {!isSearching && error ? (
        <ErrorState description={error} onRetry={() => void runSearch()} title="Search failed" />
      ) : null}

      {!isSearching && !error && !hasSearched ? (
        <EmptyState
          title="Search your saved knowledge"
          description="Try a topic like Kafka consumer groups, RAG architecture, goroutines, or system design."
        />
      ) : null}

      {!isSearching && !error && hasSearched && visibleResults.length === 0 ? (
        <EmptyState
          title="No memories found"
          description={`We couldn't find anything matching "${query.trim()}". Try different keywords, a broader question, or another Space.`}
        />
      ) : null}

      {!isSearching && !error && visibleResults.length > 0 ? (
        <section className="grid gap-3">
          <p className="text-sm text-muted-foreground">
            Showing {visibleResults.length} saved item
            {visibleResults.length === 1 ? "" : "s"} from {total} ranked match
            {total === 1 ? "" : "es"}
          </p>
          {visibleResults.map((item) => (
            <SearchResultCard item={item} key={item.content_id} />
          ))}
        </section>
      ) : null}
    </div>
  );
}

function dedupeByContent(results: SearchResult[]) {
  const byContentID = new Map<string, SearchResult>();
  for (const result of results) {
    const existing = byContentID.get(result.content_id);
    if (!existing || result.score > existing.score) {
      byContentID.set(result.content_id, result);
    }
  }
  return Array.from(byContentID.values());
}

function SearchSkeleton() {
  return (
    <section className="grid gap-3" aria-label="Search loading">
      {[0, 1, 2, 3, 4].map((item) => (
        <div className="overflow-hidden rounded-lg border bg-card shadow-sm" key={item}>
          <div className="grid gap-4 sm:grid-cols-[220px_1fr_auto]">
            <div className="aspect-video animate-pulse bg-muted" />
            <div className="space-y-3 p-4">
              <div className="h-4 w-40 animate-pulse rounded-md bg-muted" />
              <div className="h-4 w-full animate-pulse rounded-md bg-muted" />
              <div className="h-4 w-3/4 animate-pulse rounded-md bg-muted" />
            </div>
            <div className="m-4 h-4 w-20 animate-pulse rounded-md bg-muted" />
          </div>
        </div>
      ))}
    </section>
  );
}
