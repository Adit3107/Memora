import Link from "next/link";

import { TagList } from "@/components/content/tag-list";
import { contentTypeLabels } from "@/data/content";
import { mockRelevance, mockSearchSnippets } from "@/data/search";
import type { SavedContentItem } from "@/types/content";

type SearchResultCardProps = {
  item: SavedContentItem;
};

export function SearchResultCard({ item }: SearchResultCardProps) {
  const Icon = item.icon;

  return (
    <Link
      className="block rounded-md border bg-card p-4 shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={`/library/${item.slug}`}
    >
      <article className="grid gap-4 sm:grid-cols-[auto_1fr_auto]">
        <div className="flex size-10 items-center justify-center rounded-md border bg-background">
          <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
        </div>
        <div className="min-w-0 space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
              {contentTypeLabels[item.type]}
            </span>
            <span className="text-xs text-muted-foreground">{item.source}</span>
            <span className="text-xs text-muted-foreground">{item.metadata}</span>
          </div>
          <div>
            <h2 className="text-base font-semibold leading-6">{item.title}</h2>
            <p className="mt-1 text-sm leading-6 text-muted-foreground">
              {mockSearchSnippets[item.slug]}
            </p>
          </div>
          <TagList tags={item.tags} />
        </div>
        <div className="flex flex-wrap gap-2 sm:block sm:text-right">
          <p className="text-sm font-medium">{item.spaceName}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            {mockRelevance[item.slug]}
          </p>
        </div>
      </article>
    </Link>
  );
}
