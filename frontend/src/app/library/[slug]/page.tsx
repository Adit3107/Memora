import { FileText, Image, Newspaper, Video } from "lucide-react";
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

const iconForType = {
  video: Video,
  document: FileText,
  article: Newspaper,
  image: Image,
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

  return {
    id: item.id,
    slug: item.id,
    title: displayContentName(item),
    type: item.type,
    source: sourceLabel(item),
    sourceUrl: item.source_url,
    thumbnailUrl: item.thumbnail_url,
    description: item.description || "Saved to Memora and ready for retrieval.",
    metadata: metadataLabel(item),
    dateLabel: formatDate(item.created_at),
    spaceSlug: item.space_id,
    spaceName: space?.name ?? "Unknown space",
    tags: tags.map((tag) => tag.name),
    status: "Ready",
    icon: iconForType[item.type],
    detail: {
      ...template,
      extractedTitle:
        item.type === "video" ? "Transcript checkpoints" : "Extracted content",
      extractedBody:
        item.type === "video"
          ? [
              "Search results for this video include chunk numbers and timestamps when the backend has transcript timing.",
              "Use the Search page to verify semantic, keyword, and hybrid retrieval against the stored transcript chunks.",
            ]
          : [
              "Search results for this document include matching chunks and page references when the extractor provides page metadata.",
              "Use the Search page to confirm that extracted text and embeddings are reaching PostgreSQL through the Go API.",
            ],
      referenceLabel: item.type === "video" ? "Source references" : "Document references",
      references: [item.id, item.space_id],
    },
  };
}

function sourceLabel(item: BackendContent) {
  if (item.type === "video") {
    return "YouTube";
  }
  if (item.type === "document") {
    return "Document";
  }
  if (item.type === "image") {
    return "Image";
  }
  return "Article";
}

function metadataLabel(item: BackendContent) {
  if (item.type === "video") {
    return "Timestamp-aware transcript chunks";
  }
  if (item.type === "document") {
    return "Page-aware document chunks";
  }
  return "Extracted chunks";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}
