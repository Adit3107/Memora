import { ExternalLink, Play } from "lucide-react";
import Link from "next/link";

import { contentTypeLabels } from "@/data/content";
import type { SavedContentItem } from "@/types/content";

import { TagList } from "./tag-list";
import { FacebookIcon, InstagramIcon } from "@/components/ui/platform-icons";

type ContentCardProps = {
  item: SavedContentItem;
};

export function ContentCard({ item }: ContentCardProps) {
  const Icon = item.icon;
  const href = item.id ? `/app/library/${item.id}` : `/app/library/${item.slug}`;
  const thumbnail = item.thumbnailUrl ?? youtubeThumbnailFromURL(item.sourceUrl);
  const visibleTags = item.tags.slice(0, 3);
  const overflowTagCount = Math.max(item.tags.length - visibleTags.length, 0);

  return (
    <Link
      className="group block overflow-hidden rounded-lg border bg-card shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-primary/50 hover:shadow-[0_18px_50px_color-mix(in_srgb,var(--accent-glow),transparent_50%)] focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={href}
    >
      <article className="flex h-full flex-col">
        <div className="relative flex aspect-video items-center justify-center border-b bg-accent/50">
          {thumbnail ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              alt=""
              className="size-full object-cover"
              src={thumbnail}
            />
          ) : (
            <div className="flex size-20 flex-col items-center justify-center rounded-2xl border bg-card">
              <Icon className="size-7 text-primary" aria-hidden="true" />
              {item.type === "document" ? (
                <span className="mt-2 font-mono text-xs text-muted-foreground">
                  {documentExtension(item.title)}
                </span>
              ) : null}
            </div>
          )}
          {item.type === "video" ? (
            <span className="absolute bottom-3 right-3 flex size-9 items-center justify-center rounded-full bg-foreground/80 text-background">
              <Play className="size-4 fill-current" />
            </span>
          ) : null}
          {item.status && item.status !== "Ready" ? (
            <span className="absolute left-3 top-3 rounded-full border bg-card px-2 py-1 text-xs font-medium">
              {item.status}
            </span>
          ) : null}
        </div>

        <div className="flex flex-1 flex-col p-4">
          <div className="flex flex-wrap items-center gap-2">
            {item.platform === "instagram" || item.source === "Instagram Reel" ? (
              <span className="inline-flex items-center gap-1.5 rounded-md border border-pink-500/30 bg-pink-500/10 px-2 py-1 text-xs font-medium text-pink-400">
                <InstagramIcon className="size-3.5" />
                <span>Instagram Reel</span>
              </span>
            ) : item.platform === "facebook" || item.source === "Facebook Reel" ? (
              <span className="inline-flex items-center gap-1.5 rounded-md border border-blue-500/30 bg-blue-500/10 px-2 py-1 text-xs font-medium text-blue-400">
                <FacebookIcon className="size-3.5" />
                <span>Facebook Reel</span>
              </span>
            ) : (
              <>
                <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
                  {contentTypeLabels[item.type]}
                </span>
                <span className="text-xs text-muted-foreground">{item.source}</span>
              </>
            )}
            {item.creator ? (
              <span className="text-xs font-medium text-muted-foreground">
                @{item.creator.replace(/^@/, "")}
              </span>
            ) : null}
          </div>

          <h3 className="mt-3 line-clamp-2 text-base font-semibold leading-6">
            {item.title}
          </h3>
          <p className="mt-2 line-clamp-2 text-sm leading-6 text-muted-foreground">
            {item.description}
          </p>

          <div className="mt-4 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
            {item.duration ? <span>{item.duration}</span> : null}
            <span>{item.metadata}</span>
            <span>{item.dateLabel}</span>
          </div>

          <div className="mt-4 flex flex-col gap-3">
            <span className="text-sm text-muted-foreground">{item.spaceName}</span>
            <div className="flex flex-wrap items-center gap-2">
              <TagList tags={visibleTags} />
              {overflowTagCount > 0 ? (
                <span className="rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground">
                  +{overflowTagCount}
                </span>
              ) : null}
            </div>
          </div>

          {item.sourceUrl ? (
            <span className="mt-auto flex items-center gap-1 pt-4 text-xs font-medium text-muted-foreground">
              Source
              <ExternalLink className="size-3" />
            </span>
          ) : null}
        </div>
      </article>
    </Link>
  );
}

function youtubeThumbnailFromURL(url?: string) {
  if (!url) {
    return undefined;
  }

  const videoID =
    url.match(/[?&]v=([^&]+)/)?.[1] ??
    url.match(/youtu\.be\/([^?&]+)/)?.[1] ??
    url.match(/shorts\/([^?&]+)/)?.[1];

  return videoID ? `https://img.youtube.com/vi/${videoID}/hqdefault.jpg` : undefined;
}

function documentExtension(title: string) {
  const extension = title.split(".").pop();
  return extension && extension.length <= 5 ? extension.toUpperCase() : "DOC";
}
