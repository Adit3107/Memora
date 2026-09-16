"use client";

import { useMemo, useState } from "react";

import { contentTypeLabels, savedContent } from "@/data/content";
import type { ContentType, SavedContentItem } from "@/types/content";

import { ContentCard } from "./content-card";
import { EmptyState } from "./empty-state";

type FilterValue = "all" | ContentType;
type SortValue = "recent" | "title" | "space";

const filters: { label: string; value: FilterValue }[] = [
  { label: "All", value: "all" },
  { label: "Videos", value: "video" },
  { label: "Documents", value: "document" },
  { label: "Articles", value: "article" },
  { label: "Images", value: "image" },
];

function sortItems(items: SavedContentItem[], sort: SortValue) {
  return [...items].sort((a, b) => {
    if (sort === "title") {
      return a.title.localeCompare(b.title);
    }

    if (sort === "space") {
      return a.spaceName.localeCompare(b.spaceName);
    }

    return savedContent.indexOf(a) - savedContent.indexOf(b);
  });
}

export function LibraryBrowser() {
  const [activeFilter, setActiveFilter] = useState<FilterValue>("all");
  const [sort, setSort] = useState<SortValue>("recent");

  const visibleItems = useMemo(() => {
    const filtered =
      activeFilter === "all"
        ? savedContent
        : savedContent.filter((item) => item.type === activeFilter);

    return sortItems(filtered, sort);
  }, [activeFilter, sort]);

  return (
    <div className="space-y-5">
      <section className="rounded-md border bg-card p-4 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex flex-wrap gap-2" aria-label="Content filters">
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
          {activeFilter === "all" ? "items" : contentTypeLabels[activeFilter]}
        </p>
      </div>

      {visibleItems.length > 0 ? (
        <section className="grid gap-3">
          {visibleItems.map((item) => (
            <ContentCard item={item} key={item.slug} />
          ))}
        </section>
      ) : (
        <EmptyState
          title="No content matches this filter"
          description="Try another content type. Future search and metadata filters will make this library much more precise."
        />
      )}
    </div>
  );
}
