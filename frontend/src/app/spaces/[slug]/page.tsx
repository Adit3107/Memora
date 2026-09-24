"use client";

import { useUser } from "@clerk/nextjs";
import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, BookOpen, Loader2 } from "lucide-react";

import { ContentCard } from "@/components/content/content-card";
import { EmptyState } from "@/components/content/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { PageHeader } from "@/components/layout/page-header";
import { buttonVariants } from "@/components/ui/button";
import { contentTypeLabels } from "@/data/content";
import {
  displayContentName,
  listContent,
  listSpaces,
  type BackendContent,
  type BackendSpace,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import type { SavedContentItem } from "@/types/content";

type SpaceDetailPageProps = {
  params: Promise<{
    slug: string;
  }>;
};

export default function SpaceDetailPage({ params }: SpaceDetailPageProps) {
  const { user, isLoaded } = useUser();
  const [spaceID, setSpaceID] = useState("");
  const [space, setSpace] = useState<BackendSpace | null>(null);
  const [items, setItems] = useState<SavedContentItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;

    void params.then(({ slug }) => {
      if (!cancelled) setSpaceID(slug);
    });

    return () => {
      cancelled = true;
    };
  }, [params]);

  useEffect(() => {
    if (!isLoaded || !user?.id || !spaceID) return;
    let cancelled = false;
    const userID = user.id;

    async function loadSpace() {
      setIsLoading(true);
      setError("");
      try {
        const [spaces, content] = await Promise.all([
          listSpaces(userID),
          listContent(userID),
        ]);
        const currentSpace = spaces.find((candidate) => candidate.id === spaceID);

        if (!currentSpace) {
          if (!cancelled) {
            setError("This Space does not exist or is not available to your account.");
          }
          return;
        }

        if (!cancelled) {
          setSpace(currentSpace);
          setItems(
            content
              .filter((item) => item.space_id === currentSpace.id)
              .map((item) => toSavedContentItem(item, currentSpace))
          );
        }
      } catch {
        if (!cancelled) {
          setError("Could not load this Space. Check that the Go backend is running.");
        }
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    }

    void loadSpace();
    return () => {
      cancelled = true;
    };
  }, [isLoaded, spaceID, user?.id]);

  if (isLoading || !spaceID) {
    return (
      <div className="flex min-h-80 items-center justify-center gap-3 text-sm text-muted-foreground">
        <Loader2 className="size-4 animate-spin" />
        Loading Space
      </div>
    );
  }

  if (error || !space) {
    return (
      <ErrorState
        description={error || "This Space could not be found."}
        onRetry={() => window.location.reload()}
        title="Space unavailable"
      />
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          eyebrow="Space"
          title={space.name}
          description={space.description || "Your saved knowledge in one focused collection."}
        />
        <Link
          className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
          href="/app/spaces"
        >
          <ArrowLeft className="size-4" />
          Back to spaces
        </Link>
      </div>

      <section className="grid gap-4 md:grid-cols-3">
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Saved content</p>
          <p className="mt-2 text-3xl font-semibold">{items.length}</p>
        </div>
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Space owner</p>
          <p className="mt-2 text-base font-semibold">You</p>
        </div>
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Created</p>
          <p className="mt-2 text-base font-semibold">{formatDate(space.created_at)}</p>
        </div>
      </section>

      <section className="space-y-4" aria-labelledby="space-content-heading">
        <div className="flex items-center gap-3">
          <BookOpen className="size-5 text-primary" />
          <h2 className="text-lg font-semibold" id="space-content-heading">
            Content in this Space
          </h2>
        </div>
        {items.length > 0 ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {items.map((item) => (
              <ContentCard
                item={item}
                key={item.id ?? item.slug}
                onDeleted={(contentID) =>
                  setItems((current) => current.filter((entry) => entry.id !== contentID))
                }
              />
            ))}
          </div>
        ) : (
          <EmptyState
            title="This Space is empty"
            description="Save a video or document to this Space and it will appear here."
          />
        )}
      </section>
    </div>
  );
}

function toSavedContentItem(item: BackendContent, space: BackendSpace): SavedContentItem {
  return {
    id: item.id,
    slug: item.id,
    title: displayContentName(item),
    type: item.type,
    source: item.type === "video" ? "Video" : contentTypeLabels[item.type],
    sourceUrl: item.source_url,
    thumbnailUrl: item.thumbnail_url,
    description: item.description || "Saved to this Space and ready for search.",
    metadata: item.type === "video" ? "Transcript chunks" : "Extracted document text",
    dateLabel: formatDate(item.created_at),
    spaceSlug: space.id,
    spaceName: space.name,
    tags: [],
    status: "Ready",
    detail: {
      heroLabel: "Saved content",
      previewTitle: displayContentName(item),
      previewBody: "This content is saved in your Space.",
      extractedTitle: item.type === "video" ? "AI Video Summary" : "AI Document Summary",
      extractedBody: [item.description || "Summary available from the content detail page."],
      referenceLabel: "Source",
      references: [item.source_url || "Local document"],
    },
  };
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}
