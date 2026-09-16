import Link from "next/link";
import { notFound } from "next/navigation";

import { ContentCard } from "@/components/content/content-card";
import { EmptyState } from "@/components/content/empty-state";
import { TagList } from "@/components/content/tag-list";
import { PageHeader } from "@/components/layout/page-header";
import { buttonVariants } from "@/components/ui/button";
import {
  contentTypeLabels,
  getContentBySpace,
  getSpaceBySlug,
  spaces,
} from "@/data/content";
import { cn } from "@/lib/utils";
import type { ContentType } from "@/types/content";

type SpaceDetailPageProps = {
  params: Promise<{
    slug: string;
  }>;
};

const orderedTypes: ContentType[] = ["video", "document", "article", "image"];

export function generateStaticParams() {
  return spaces.map((space) => ({ slug: space.slug }));
}

export default async function SpaceDetailPage({ params }: SpaceDetailPageProps) {
  const { slug } = await params;
  const space = getSpaceBySlug(slug);

  if (!space) {
    notFound();
  }

  const items = getContentBySpace(space.slug);

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          eyebrow="Space"
          title={space.name}
          description={space.description}
        />
        <Link
          className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
          href="/spaces"
        >
          Back to spaces
        </Link>
      </div>

      <section className="grid gap-4 md:grid-cols-3">
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Saved content</p>
          <p className="mt-2 text-3xl font-semibold">{items.length}</p>
        </div>
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Owner</p>
          <p className="mt-2 text-base font-semibold">{space.ownerLabel}</p>
        </div>
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Activity</p>
          <p className="mt-2 text-base font-semibold">{space.updatedAt}</p>
        </div>
      </section>

      <section className="rounded-md border bg-card p-5 shadow-sm">
        <h2 className="text-lg font-semibold">Space tags</h2>
        <div className="mt-3">
          <TagList tags={space.tags} />
        </div>
      </section>

      <section className="rounded-md border bg-card p-5 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 className="text-lg font-semibold">Future Space workflow</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              This view is shaped for browsing saved formats before future
              Search and Ask AI capabilities are connected.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            {orderedTypes.map((type) => (
              <span
                className="rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground"
                key={type}
              >
                {contentTypeLabels[type]}
              </span>
            ))}
            <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
              Search
            </span>
            <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
              Ask AI later
            </span>
          </div>
        </div>
      </section>

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Content in this Space</h2>
        {items.length > 0 ? (
          <div className="grid gap-3">
            {items.map((item) => (
              <ContentCard item={item} key={item.slug} />
            ))}
          </div>
        ) : (
          <EmptyState
            title="This Space is empty"
            description="When content is saved to this Space, videos, documents, articles, and images will appear here."
          />
        )}
      </section>
    </div>
  );
}
