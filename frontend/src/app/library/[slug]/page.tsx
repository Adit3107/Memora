import { notFound } from "next/navigation";

import { ContentDetail } from "@/components/content/content-detail";
import { getContentBySlug, savedContent } from "@/data/content";
import {
  displayContentName,
  getContent,
  listContentTags,
  listSpaces,
  type BackendContent,
  type BackendSpace,
  type BackendTag,
} from "@/lib/api";
import type { SavedContentItem } from "@/types/content";

type ContentDetailPageProps = {
  params: Promise<{
    slug: string;
  }>;
};



export default async function ContentDetailPage({
  params,
}: ContentDetailPageProps) {
  const { slug } = await params;
  let item: SavedContentItem | undefined;

  try {
    const [content, spaces, tags] = await Promise.all([
      getContent(slug),
      listSpaces(),
      listContentTags(slug).catch(() => []),
    ]);
    item = toSavedContentItem(content, spaces, tags);
  } catch {
    item = getContentBySlug(slug);
  }

  if (!item) {
    notFound();
  }

  return <ContentDetail item={item} />;
}

function toSavedContentItem(
  item: BackendContent,
  spaces: BackendSpace[],
  tags: BackendTag[]
): SavedContentItem {
  const space = spaces.find((candidate) => candidate.id === item.space_id);
  const template =
    savedContent.find((mock) => mock.type === item.type)?.detail ??
    savedContent[0].detail;

  const platform = detectPlatform(item.source_url);
  const meta = extractPlatformMetadata(item, platform);

  const pipeline =
    item.type === "video"
      ? platform === "youtube"
        ? "yt-dlp + Faster-Whisper Audio"
        : "Whisper Audio + RapidOCR Vision"
      : "Document Parser & Chunker";

  return {
    id: item.id,
    slug: item.id,
    title: displayContentName(item),
    type: item.type,
    source: sourceLabel(item, platform),
    sourceUrl: item.source_url,
    thumbnailUrl: item.thumbnail_url,
    description: item.description || "Saved to Mindshelf and ready for retrieval.",
    metadata: meta,
    dateLabel: formatDate(item.created_at),
    spaceSlug: item.space_id,
    spaceName: space?.name ?? "Default Space",
    tags: tags.map((tag) => tag.name),
    status: "Ready",
    platform,
    detail: {
      ...template,
      extractedTitle:
        item.type === "video" ? "AI Video Summary" : "AI Document Summary",
      extractedBody: [
        item.description || "Extracting key takeaways with Gemini AI...",
      ],
      referenceLabel: "Source & Provenance",
      references: [
        `Source: ${platform ? platform.toUpperCase() : "Document"}`,
        `Space: ${space?.name ?? "General"}`,
        `Pipeline: ${pipeline}`,
        `Storage: PostgreSQL + pgvector`,
      ],
    },
  };
}

function detectPlatform(url?: string): "instagram" | "facebook" | "youtube" | undefined {
  if (!url) return undefined;
  const lower = url.toLowerCase();
  if (lower.includes("instagram.com") || lower.includes("instagr.am")) return "instagram";
  if (lower.includes("facebook.com") || lower.includes("fb.watch") || lower.includes("fb.com"))
    return "facebook";
  if (lower.includes("youtube.com") || lower.includes("youtu.be")) return "youtube";
  return undefined;
}

function extractPlatformMetadata(
  item: BackendContent,
  platform?: "instagram" | "facebook" | "youtube"
): string {
  const url = item.source_url || "";

  if (platform === "youtube") {
    // For YouTube: Show the title
    return item.title ? `Title: ${item.title}` : "YouTube Video";
  }

  if (platform === "instagram") {
    // For Instagram: Show the Reel ID
    const match = url.match(/\/(?:reel|p)\/([A-Za-z0-9_-]+)/i);
    const reelId = match ? match[1] : "Reel";
    return `Reel ID: ${reelId}`;
  }

  if (platform === "facebook") {
    // For Facebook: Show the Reel/Video ID
    const match =
      url.match(/\/(?:reel|videos|watch\/?\?v=)([0-9]+)/i) ||
      url.match(/[?&]v=([0-9]+)/i) ||
      url.match(/\/reel\/([0-9a-zA-Z_-]+)/i);
    const reelId = match ? match[1] : "Reel";
    return `Reel ID: ${reelId}`;
  }

  return item.title || item.name || "Document";
}

function sourceLabel(
  item: BackendContent,
  platform?: "instagram" | "facebook" | "youtube"
) {
  if (item.type === "video") {
    if (platform === "instagram") return "Instagram Reel";
    if (platform === "facebook") return "Facebook Reel";
    if (platform === "youtube") return "YouTube Video";
    return "Video";
  }
  if (item.type === "document") {
    return "Document";
  }
  if (item.type === "image") {
    return "Image";
  }
  return "Article";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}
