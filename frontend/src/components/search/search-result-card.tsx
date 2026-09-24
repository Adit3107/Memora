import Link from "next/link";
import { Clock, FileText, Image, Newspaper, Video } from "lucide-react";

import { TagList } from "@/components/content/tag-list";
import { contentTypeLabels } from "@/data/content";
import { FacebookIcon, InstagramIcon } from "@/components/ui/platform-icons";
import type { SearchResult } from "@/lib/api";

type SearchResultCardProps = {
  item: SearchResult;
};

export function SearchResultCard({ item }: SearchResultCardProps) {
  const Icon = iconForType[item.content_type] || Video;
  const href = `/app/library/${item.content_id}`;
  const thumbnail = item.thumbnail_url || youtubeThumbnailFromURL(item.source_url);
  const timeLabel = formatTimestamp(item.start_seconds, item.end_seconds);
  const platform = detectPlatformForResult(item);
  const creator = item.metadata?.uploader;

  return (
    <Link
      className="block rounded-md border bg-card p-4 shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={href}
    >
      <article className="grid gap-4 sm:grid-cols-[auto_1fr_auto]">
        {thumbnail ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            alt=""
            className="h-16 w-24 rounded-md border object-cover"
            src={thumbnail}
          />
        ) : (
          <div className="flex size-10 items-center justify-center rounded-md border bg-background">
            <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
          </div>
        )}
        <div className="min-w-0 space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            {platform === "instagram" ? (
              <span className="inline-flex items-center gap-1.5 rounded-md border border-pink-500/30 bg-pink-500/10 px-2 py-0.5 text-xs font-medium text-pink-400">
                <InstagramIcon className="size-3" />
                <span>Instagram Reel</span>
              </span>
            ) : platform === "facebook" ? (
              <span className="inline-flex items-center gap-1.5 rounded-md border border-blue-500/30 bg-blue-500/10 px-2 py-0.5 text-xs font-medium text-blue-400">
                <FacebookIcon className="size-3" />
                <span>Facebook Reel</span>
              </span>
            ) : (
              <>
                <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
                  {contentTypeLabels[item.content_type]}
                </span>
                <span className="text-xs text-muted-foreground">{sourceLabelForResult(item)}</span>
              </>
            )}

            {creator ? (
              <span className="text-xs font-medium text-muted-foreground">
                @{creator.replace(/^@/, "")}
              </span>
            ) : null}

            {timeLabel ? (
              <span className="inline-flex items-center gap-1 rounded-md border bg-muted/60 px-2 py-0.5 font-mono text-xs text-muted-foreground">
                <Clock className="size-3" />
                <span>{timeLabel}</span>
              </span>
            ) : null}
          </div>
          <div>
            <h2 className="text-base font-semibold leading-6">{item.title}</h2>
            <p className="mt-1 text-sm leading-6 text-muted-foreground">
              {item.text}
            </p>
          </div>
          <TagList tags={item.tags} />
        </div>
        <div className="flex flex-wrap gap-2 sm:block sm:text-right">
          <p className="text-sm font-medium">{scoreLabel(item.score)}</p>
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

function detectPlatformForResult(item: SearchResult): "instagram" | "facebook" | "youtube" | null {
  if (
    item.metadata?.provider === "instagram" ||
    item.metadata?.source_platform === "instagram" ||
    item.source_url?.includes("instagram.com")
  ) {
    return "instagram";
  }
  if (
    item.metadata?.provider === "facebook" ||
    item.metadata?.source_platform === "facebook" ||
    item.source_url?.includes("facebook.com") ||
    item.source_url?.includes("fb.watch")
  ) {
    return "facebook";
  }
  if (item.source_url?.includes("youtube.com") || item.source_url?.includes("youtu.be")) {
    return "youtube";
  }
  return null;
}

function sourceLabelForResult(item: SearchResult) {
  if (item.source_type === "reel") {
    return "Reel";
  }
  if (item.source_type === "video") {
    return "Video";
  }
  if (item.source_type === "document") {
    return "Document";
  }
  return item.source_type;
}

function formatTimestamp(start?: number, end?: number): string | null {
  if (start === undefined || start === null) {
    return null;
  }
  const formatSec = (s: number) => {
    const mins = Math.floor(s / 60);
    const secs = Math.floor(s % 60);
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };
  if (end !== undefined && end !== null && end > start) {
    return `${formatSec(start)} - ${formatSec(end)}`;
  }
  return formatSec(start);
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
