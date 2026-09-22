"use client";

import { Search } from "lucide-react";
import { FormEvent, useState } from "react";

import { EmptyState } from "@/components/content/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { searchMemora, type SearchMode, type SearchResult } from "@/lib/api";
import type { ContentType } from "@/types/content";

import { SearchResultCard } from "./search-result-card";

const defaultUserID =
  process.env.NEXT_PUBLIC_MEMORA_USER_ID ??
  "67a79aff-376d-48a7-af69-f087d46d313e";
const defaultSpaceID =
  process.env.NEXT_PUBLIC_MEMORA_SPACE_ID ??
  "2e96706f-8134-448b-918d-979aeb0500bc";

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

export function SearchBrowser() {
  const [query, setQuery] = useState("");
  const [userID, setUserID] = useState(defaultUserID);
  const [spaceID, setSpaceID] = useState(defaultSpaceID);
  const [mode, setMode] = useState<SearchMode>("hybrid");
  const [contentType, setContentType] = useState<"all" | ContentType>("all");
  const [sourceType, setSourceType] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState("");

  async function runSearch(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();
    const trimmedQuery = query.trim();
    if (!trimmedQuery) {
      setResults([]);
      setHasSearched(false);
      setError("");
      return;
    }

    setIsSearching(true);
    setError("");
    try {
      const response = await searchMemora({
        user_id: userID.trim(),
        space_id: spaceID.trim() || undefined,
        query: trimmedQuery,
        mode,
        content_type: contentType === "all" ? undefined : contentType,
        source_type: sourceType.trim() || undefined,
        limit: 10,
        offset: 0,
      });
      setResults(response.results);
      setHasSearched(true);
    } catch {
      setResults([]);
      setHasSearched(true);
      setError("Search failed. Check that the backend, AI service, and database are running.");
    } finally {
      setIsSearching(false);
    }
  }

  return (
    <div className="space-y-6">
      <form className="space-y-4 rounded-md border bg-card p-5 shadow-sm" onSubmit={runSearch}>
        <label className="text-sm font-medium" htmlFor="memora-search">
          Natural-language search
        </label>
        <div className="mt-3 flex items-center gap-3 rounded-md border bg-background px-3 py-2">
          <Search className="size-4 text-muted-foreground" aria-hidden="true" />
          <input
            className="h-10 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            id="memora-search"
            onChange={(event) => setQuery(event.target.value)}
            placeholder="How do Kafka consumer groups distribute partitions?"
            type="search"
            value={query}
          />
        </div>

        <div className="grid gap-3 md:grid-cols-2">
          <label className="space-y-2 text-sm font-medium">
            User ID
            <input
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setUserID(event.target.value)}
              value={userID}
            />
          </label>
          <label className="space-y-2 text-sm font-medium">
            Space ID
            <input
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setSpaceID(event.target.value)}
              value={spaceID}
            />
          </label>
        </div>

        <div className="grid gap-3 md:grid-cols-3">
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
            Content type
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
            Source type
            <input
              className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
              onChange={(event) => setSourceType(event.target.value)}
              placeholder="video, document, image"
              value={sourceType}
            />
          </label>
        </div>

        <button
          className="h-10 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={isSearching || !query.trim() || !userID.trim()}
          type="submit"
        >
          Search
        </button>
      </form>

      {isSearching ? (
        <LoadingState title="Searching Memora" description="Retrieving matching chunks." />
      ) : null}

      {!isSearching && error ? (
        <ErrorState description={error} onRetry={() => void runSearch()} title="Search failed" />
      ) : null}

      {!isSearching && !error && !hasSearched ? (
        <EmptyState
          title="Search your saved knowledge"
          description="Enter a query to search vectorized YouTube transcripts and document chunks."
        />
      ) : null}

      {!isSearching && !error && hasSearched && results.length === 0 ? (
        <EmptyState
          title="No matching chunks"
          description="No saved content matched this query and filter combination."
        />
      ) : null}

      {!isSearching && !error && results.length > 0 ? (
        <section className="grid gap-3">
          <p className="text-sm text-muted-foreground">
            Showing {results.length} ranked results
          </p>
          {results.map((item) => (
            <SearchResultCard item={item} key={item.chunk_id} />
          ))}
        </section>
      ) : null}
    </div>
  );
}
