import Link from "next/link";

import { contentTypeLabels } from "@/data/content";
import type { SavedContentItem } from "@/types/content";

import { TagList } from "./tag-list";

type ContentCardProps = {
  item: SavedContentItem;
};

export function ContentCard({ item }: ContentCardProps) {
  const Icon = item.icon;

  return (
    <Link
      className="block rounded-md border bg-card p-4 shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={`/library/${item.slug}`}
    >
      <article className="grid gap-4 sm:grid-cols-[auto_1fr]">
        <div className="flex size-10 items-center justify-center rounded-md border bg-background">
          <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
        </div>
        <div className="min-w-0 space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
              {contentTypeLabels[item.type]}
            </span>
            <span className="text-xs text-muted-foreground">{item.source}</span>
            <span className="text-xs text-muted-foreground">{item.dateLabel}</span>
          </div>
          <div>
            <h3 className="text-base font-semibold leading-6">{item.title}</h3>
            <p className="mt-1 text-sm leading-6 text-muted-foreground">
              {item.description}
            </p>
          </div>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <span className="text-sm text-muted-foreground">
              Space: {item.spaceName}
            </span>
            <TagList tags={item.tags} />
          </div>
        </div>
      </article>
    </Link>
  );
}
