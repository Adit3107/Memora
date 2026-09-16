"use client";

import { Search } from "lucide-react";
import { useMemo, useState } from "react";

import { EmptyState } from "@/components/content/empty-state";
import { savedContent, spaces } from "@/data/content";
import { allTags, defaultSearchQuery } from "@/data/search";
import type { ContentType } from "@/types/content";
import type { SearchFilter } from "@/types/search";

import { SearchResultCard } from "./search-result-card";

const contentTypeOptions: { label: string; value: "all" | ContentType }[] = [
  { label: "All", value: "all" },
  { label: "Videos", value: "video" },
  { label: "Documents", value: "document" },
  { label: "Articles", value: "article" },
  { label: "Images", value: "image" },
];

function matchesQuery(itemText: string, query: string) {
  const terms = query
    .toLowerCase()
    .split(/\s+/)
    .map((term) => term.trim())
    .filter((term) => term.length > 2);

  if (terms.length === 0) {
    return true;
  }

  return terms.some((term) => itemText.includes(term));
}

export function SearchBrowser() {
  const [query, setQuery] = useState(defaultSearchQuery);
  const [filters, setFilters] = useState<SearchFilter>({
    contentType: "all",
    spaceSlug: "all",
    tag: "all",
  });

  const results = useMemo(() => {
    const normalizedQuery = query.toLowerCase();

    return savedContent.filter((item) => {
      const searchableText = [
        item.title,
        item.description,
        item.source,
        item.spaceName,
        item.tags.join(" "),
        item.detail.extractedBody.join(" "),
      ]
        .join(" ")
        .toLowerCase();

      const typeMatch =
        filters.contentType === "all" || item.type === filters.contentType;
      const spaceMatch =
        filters.spaceSlug === "all" || item.spaceSlug === filters.spaceSlug;
      const tagMatch = filters.tag === "all" || item.tags.includes(filters.tag);

      return (
        typeMatch &&
        spaceMatch &&
        tagMatch &&
        matchesQuery(searchableText, normalizedQuery)
      );
    });
  }, [filters, query]);

  return (
    <div className="space-y-6">
      <section className="rounded-md border bg-card p-5 shadow-sm">
        <label className="text-sm font-medium" htmlFor="memora-search">
          Natural-language search
        </label>
        <div className="mt-3 flex items-center gap-3 rounded-md border bg-background px-3 py-2">
          <Search className="size-4 text-muted-foreground" aria-hidden="true" />
          <input
            className="h-10 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            id="memora-search"
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Where did I save the video explaining vector databases?"
            type="search"
            value={query}
          />
        </div>
        <p className="mt-3 text-sm leading-6 text-muted-foreground">
          Mock matching only. Semantic search, embeddings, pgvector, and
          reranking belong to later phases.
        </p>
      </section>

      <section className="grid gap-3 rounded-md border bg-card p-4 shadow-sm md:grid-cols-3">
        <label className="space-y-2 text-sm font-medium">
          Content type
          <select
            className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
            onChange={(event) =>
              setFilters((current) => ({
                ...current,
                contentType: event.target.value as SearchFilter["contentType"],
              }))
            }
            value={filters.contentType}
          >
            {contentTypeOptions.map((option) => (
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
            onChange={(event) =>
              setFilters((current) => ({
                ...current,
                spaceSlug: event.target.value,
              }))
            }
            value={filters.spaceSlug}
          >
            <option value="all">All Spaces</option>
            {spaces.map((space) => (
              <option key={space.slug} value={space.slug}>
                {space.name}
              </option>
            ))}
          </select>
        </label>

        <label className="space-y-2 text-sm font-medium">
          Tag
          <select
            className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
            onChange={(event) =>
              setFilters((current) => ({
                ...current,
                tag: event.target.value,
              }))
            }
            value={filters.tag}
          >
            <option value="all">All tags</option>
            {allTags.map((tag) => (
              <option key={tag} value={tag}>
                #{tag}
              </option>
            ))}
          </select>
        </label>
      </section>

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-muted-foreground">
          Showing {results.length} mock results
        </p>
        <p className="text-xs text-muted-foreground">
          Result cards are prepared for snippets, timestamps, pages, Spaces, and
          tags.
        </p>
      </div>

      {results.length > 0 ? (
        <section className="grid gap-3">
          {results.map((item) => (
            <SearchResultCard item={item} key={item.slug} />
          ))}
        </section>
      ) : (
        <EmptyState
          title="No mock results"
          description={`No saved mock content matches "${query}". Try clearing filters or searching for RAG, vector, Kafka, Go, or DSA.`}
        />
      )}
    </div>
  );
}
