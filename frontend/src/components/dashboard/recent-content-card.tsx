import type { RecentContentItem } from "@/types/dashboard";

type RecentContentCardProps = {
  item: RecentContentItem;
};

export function RecentContentCard({ item }: RecentContentCardProps) {
  const Icon = item.icon;

  return (
    <article className="grid gap-4 rounded-md border bg-card p-4 shadow-sm sm:grid-cols-[auto_1fr_auto] sm:items-start">
      <div className="flex size-10 items-center justify-center rounded-md border bg-background">
        <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
      </div>

      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <span className="rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground">
            {item.type}
          </span>
          <span className="text-xs text-muted-foreground">{item.source}</span>
        </div>
        <h3 className="mt-3 text-base font-semibold leading-6">{item.title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{item.metadata}</p>
      </div>

      <div className="flex flex-wrap gap-2 sm:justify-end">
        <span className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground">
          {item.space}
        </span>
        <span className="rounded-md border px-2 py-1 text-xs text-muted-foreground">
          #{item.tag}
        </span>
      </div>
    </article>
  );
}
