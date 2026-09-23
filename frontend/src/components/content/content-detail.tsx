import Link from "next/link";

import { contentTypeLabels } from "@/data/content";
import { cn } from "@/lib/utils";
import type { SavedContentItem } from "@/types/content";

import { buttonVariants } from "../ui/button";
import { TagList } from "./tag-list";

type ContentDetailProps = {
  item: SavedContentItem;
};

export function ContentDetail({ item }: ContentDetailProps) {
  const Icon = item.icon;

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="max-w-3xl space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
              {contentTypeLabels[item.type]}
            </span>
            <span className="text-xs text-muted-foreground">{item.source}</span>
            <span className="text-xs text-muted-foreground">{item.dateLabel}</span>
          </div>
          <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl">
            {item.title}
          </h1>
          <p className="text-base leading-7 text-muted-foreground">
            {item.description}
          </p>
        </div>
        <Link
          className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
          href="/app/library"
        >
          Back to library
        </Link>
      </div>

      <section className="grid gap-4 md:grid-cols-3">
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Space</p>
          <Link
            className="mt-2 block text-base font-semibold hover:underline"
            href={`/app/spaces/${item.spaceSlug}`}
          >
            {item.spaceName}
          </Link>
        </div>
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Metadata</p>
          <p className="mt-2 text-base font-semibold">{item.metadata}</p>
        </div>
        <div className="rounded-md border bg-card p-4 shadow-sm">
          <p className="text-sm text-muted-foreground">Source</p>
          {item.sourceUrl ? (
            <Link
              className="mt-2 block break-all text-base font-semibold hover:underline"
              href={item.sourceUrl}
            >
              {item.source}
            </Link>
          ) : (
            <p className="mt-2 text-base font-semibold">{item.source}</p>
          )}
        </div>
      </section>

      <section className="rounded-md border bg-card p-5 shadow-sm">
        <h2 className="text-lg font-semibold">Tags</h2>
        <div className="mt-3">
          <TagList tags={item.tags} />
        </div>
      </section>

      <div className="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
        <section className="rounded-md border bg-card p-5 shadow-sm">
          <div className="flex min-h-64 flex-col items-center justify-center rounded-md border border-dashed bg-background p-6 text-center">
            {item.thumbnailUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                alt=""
                className="mb-4 aspect-video w-full max-w-md rounded-md border object-cover"
                src={item.thumbnailUrl}
              />
            ) : (
              <div className="flex size-12 items-center justify-center rounded-md border bg-card">
                <Icon className="size-6 text-muted-foreground" aria-hidden="true" />
              </div>
            )}
            <p className="mt-4 text-sm font-medium">{item.detail.heroLabel}</p>
            <h2 className="mt-2 text-xl font-semibold">
              {item.detail.previewTitle}
            </h2>
            <p className="mt-2 max-w-md text-sm leading-6 text-muted-foreground">
              {item.detail.previewBody}
            </p>
          </div>
        </section>

        <section className="rounded-md border bg-card p-5 shadow-sm">
          <h2 className="text-lg font-semibold">{item.detail.extractedTitle}</h2>
          <div className="mt-4 divide-y">
            {item.detail.extractedBody.map((line) => (
              <p className="py-3 text-sm leading-6 text-muted-foreground" key={line}>
                {line}
              </p>
            ))}
          </div>
        </section>
      </div>

      <section className="rounded-md border bg-card p-5 shadow-sm">
        <h2 className="text-lg font-semibold">{item.detail.referenceLabel}</h2>
        <div className="mt-4 flex flex-wrap gap-2">
          {item.detail.references.map((reference) => (
            <span
              className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground"
              key={reference}
            >
              {reference}
            </span>
          ))}
        </div>
      </section>
    </div>
  );
}
