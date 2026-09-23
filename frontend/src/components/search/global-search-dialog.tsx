"use client";

import { Search, X } from "lucide-react";
import Link from "next/link";
import { FormEvent, useEffect, useRef, useState } from "react";

import {
  MEMORA_DEMO_USER_ID,
  searchMemora,
  type SearchResult,
} from "@/lib/api";

type GlobalSearchDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function GlobalSearchDialog({
  open,
  onOpenChange,
}: GlobalSearchDialogProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }

    const timeout = window.setTimeout(() => inputRef.current?.focus(), 60);
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onOpenChange(false);
      }
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.clearTimeout(timeout);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [onOpenChange, open]);

  async function runSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = query.trim();
    if (!trimmed) {
      return;
    }

    setIsSearching(true);
    setError("");
    try {
      const response = await searchMemora({
        user_id: MEMORA_DEMO_USER_ID,
        query: trimmed,
        mode: "hybrid",
        limit: 6,
        offset: 0,
      });
      setResults(response.results);
    } catch {
      setResults([]);
      setError("Search could not be completed. Check that the backend and AI service are running.");
    } finally {
      setIsSearching(false);
    }
  }

  if (!open) {
    return null;
  }

  return (
    <div
      aria-modal="true"
      className="fixed inset-0 z-50 bg-foreground/20 px-4 py-16 backdrop-blur-sm"
      role="dialog"
    >
      <button
        aria-label="Close search"
        className="absolute inset-0 size-full cursor-default"
        onClick={() => onOpenChange(false)}
        type="button"
      />
      <div className="relative mx-auto max-w-2xl overflow-hidden rounded-2xl border bg-card shadow-2xl">
        <div className="flex items-center justify-between border-b px-4 py-3">
          <div>
            <h2 className="text-sm font-semibold">Search your memory</h2>
            <p className="text-xs text-muted-foreground">Semantic, keyword, and hybrid retrieval.</p>
          </div>
          <button
            aria-label="Close search"
            className="inline-flex size-8 items-center justify-center rounded-md border bg-background"
            onClick={() => onOpenChange(false)}
            type="button"
          >
            <X className="size-4" />
          </button>
        </div>

        <form className="border-b p-4" onSubmit={runSearch}>
          <div className="flex items-center gap-3 rounded-md border bg-background px-3 py-2">
            <Search className="size-4 text-muted-foreground" />
            <input
              className="h-10 flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              onChange={(event) => setQuery(event.target.value)}
              placeholder="kafka consumer groups"
              ref={inputRef}
              value={query}
            />
            <button
              className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-50"
              disabled={isSearching || !query.trim()}
              type="submit"
            >
              {isSearching ? "Searching" : "Search"}
            </button>
          </div>
        </form>

        <div className="max-h-[55vh] overflow-y-auto p-3">
          {isSearching ? <SearchDialogSkeleton /> : null}
          {!isSearching && error ? (
            <p className="rounded-md border bg-background p-3 text-sm text-muted-foreground">
              {error}
            </p>
          ) : null}
          {!isSearching && !error && results.length === 0 ? (
            <p className="rounded-md border bg-background p-3 text-sm text-muted-foreground">
              Type a query to find saved videos, documents, articles, and images.
            </p>
          ) : null}
          {!isSearching && !error && results.length > 0 ? (
            <div className="space-y-2">
              {results.map((result) => (
                <Link
                  className="block rounded-md border bg-background p-3 transition-colors hover:bg-accent/60"
                  href={`/app/library/${result.content_id}?chunk=${result.chunk_id}`}
                  key={result.chunk_id}
                  onClick={() => onOpenChange(false)}
                >
                  <p className="text-sm font-semibold">{result.title}</p>
                  <p className="mt-1 line-clamp-2 text-sm leading-6 text-muted-foreground">
                    {result.text}
                  </p>
                  <p className="mt-2 text-xs text-muted-foreground">
                    {result.source_type} - chunk {result.chunk_index + 1}
                  </p>
                </Link>
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}

function SearchDialogSkeleton() {
  return (
    <div className="space-y-2">
      {[0, 1, 2].map((item) => (
        <div className="rounded-md border bg-background p-3" key={item}>
          <div className="h-4 w-1/2 animate-pulse rounded-md bg-muted" />
          <div className="mt-3 h-4 w-full animate-pulse rounded-md bg-muted" />
          <div className="mt-2 h-4 w-2/3 animate-pulse rounded-md bg-muted" />
        </div>
      ))}
    </div>
  );
}
