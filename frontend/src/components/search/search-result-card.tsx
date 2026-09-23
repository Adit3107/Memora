import Link from "next/link";
import { FileText, Image, Newspaper, Video } from "lucide-react";

import { TagList } from "@/components/content/tag-list";
import { contentTypeLabels } from "@/data/content";
import type { SearchResult } from "@/lib/api";

type SearchResultCardProps = {
  item: SearchResult;
};

export function SearchResultCard({ item }: SearchResultCardProps) {
  const Icon = iconForType[item.content_type];
  const href = `/app/library/${item.content_id}?chunk=${item.chunk_id}`;
  const sourceLabel = sourceLabelForResult(item);

  return (
    <Link
      className="block rounded-md border bg-card p-4 shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={href}
    >
      <article className="grid gap-4 sm:grid-cols-[auto_1fr_auto]">
        {item.thumbnail_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            alt=""
            className="h-16 w-24 rounded-md border object-cover"
            src={item.thumbnail_url}
          />
        ) : (
          <div className="flex size-10 items-center justify-center rounded-md border bg-background">
            <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
          </div>
        )}
        <div className="min-w-0 space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
              {contentTypeLabels[item.content_type]}
            </span>
            <span className="text-xs text-muted-foreground">{sourceLabel}</span>
            <span className="text-xs text-muted-foreground">
              chunk {item.chunk_index + 1}
            </span>
          </div>
          <div>
            <h2 className="text-base font-semibold leading-6">{item.title}</h2>
            <p className="mt-1 text-sm leading-6 text-muted-foreground">
              {item.text}
            </p>
            <p className="mt-2 text-xs leading-5 text-muted-foreground">
              Returned from {sourceReference(item)} because this stored chunk matched
              the search query.
            </p>
          </div>
          <TagList tags={item.tags} />
        </div>
        <div className="flex flex-wrap gap-2 sm:block sm:text-right">
          <p className="text-sm font-medium">{scoreLabel(item.score)}</p>
          <p className="mt-1 text-xs text-muted-foreground">
            {sourceReference(item)}
          </p>
        </div>
      </article>
    </Link>
  );
}

const iconForType = {
  video: Video,
  document: FileText,
  article: Newspaper,
  image: Image,
};

function scoreLabel(score: number) {
  return `${Math.round(score * 100)}% match`;
}

function sourceReference(item: SearchResult) {
  if (item.start_seconds !== undefined) {
    return `starts at ${formatSeconds(item.start_seconds)}`;
  }
  if (item.page_index !== undefined) {
    return `page ${item.page_index}`;
  }
  return item.source_type;
}

function sourceLabelForResult(item: SearchResult) {
  if (item.source_type === "video") {
    return "Video";
  }
  if (item.source_type === "document") {
    return "Document";
  }
  return item.source_type;
}

function formatSeconds(seconds: number) {
  const minutes = Math.floor(seconds / 60);
  const remaining = Math.floor(seconds % 60);
  return `${minutes}:${remaining.toString().padStart(2, "0")}`;
}
