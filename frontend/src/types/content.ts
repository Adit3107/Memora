export type ContentType = "video" | "document" | "article" | "image";

export type SavedContentItem = {
  id?: string;
  slug: string;
  title: string;
  type: ContentType;
  source: string;
  sourceUrl?: string;
  thumbnailUrl?: string;
  description: string;
  metadata: string;
  dateLabel: string;
  spaceSlug: string;
  spaceName: string;
  tags: string[];
  status?: string;
  creator?: string;
  duration?: string;
  platform?: "youtube" | "instagram" | "facebook" | string;
  detail: {
    heroLabel: string;
    previewTitle: string;
    previewBody: string;
    extractedTitle: string;
    extractedBody: string[];
    referenceLabel: string;
    references: string[];
  };
};

export type Space = {
  slug: string;
  name: string;
  description: string;
  ownerLabel: string;
  updatedAt: string;
  tags: string[];
};
