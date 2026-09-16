import type { ContentTypeSummary } from "@/types/dashboard";

type ContentTypeCardProps = {
  item: ContentTypeSummary;
};

export function ContentTypeCard({ item }: ContentTypeCardProps) {
  const Icon = item.icon;

  return (
    <article className="rounded-md border bg-card p-4 shadow-sm">
      <div className="flex items-center gap-3">
        <div className="flex size-9 items-center justify-center rounded-md border bg-background">
          <Icon className="size-4 text-muted-foreground" aria-hidden="true" />
        </div>
        <div>
          <h3 className="text-sm font-semibold">{item.type}</h3>
          <p className="text-xs text-muted-foreground">{item.count} items</p>
        </div>
      </div>
      <p className="mt-4 text-sm leading-6 text-muted-foreground">
        {item.detail}
      </p>
    </article>
  );
}
