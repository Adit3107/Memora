import Link from "next/link";

import { getContentBySpace } from "@/data/content";
import type { Space } from "@/types/content";

import { TagList } from "../content/tag-list";

type SpaceCardProps = {
  space: Space;
};

export function SpaceCard({ space }: SpaceCardProps) {
  const itemCount = getContentBySpace(space.slug).length;

  return (
    <Link
      className="block rounded-md border bg-card p-5 shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={`/spaces/${space.slug}`}
    >
      <article className="space-y-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">{space.name}</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              {space.description}
            </p>
          </div>
          <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
            {itemCount} items
          </span>
        </div>
        <div className="grid gap-2 text-sm text-muted-foreground sm:grid-cols-2">
          <span>{space.ownerLabel}</span>
          <span className="sm:text-right">{space.updatedAt}</span>
        </div>
        <TagList tags={space.tags} />
      </article>
    </Link>
  );
}
