import type { DashboardOverviewItem } from "@/types/dashboard";

type OverviewCardProps = {
  item: DashboardOverviewItem;
};

export function OverviewCard({ item }: OverviewCardProps) {
  const Icon = item.icon;

  return (
    <article className="rounded-md border bg-card p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground">{item.label}</p>
          <p className="mt-3 text-3xl font-semibold tracking-normal">
            {item.value}
          </p>
        </div>
        <div className="flex size-9 items-center justify-center rounded-md border bg-background">
          <Icon className="size-4 text-muted-foreground" aria-hidden="true" />
        </div>
      </div>
      <p className="mt-3 text-sm leading-6 text-muted-foreground">
        {item.detail}
      </p>
    </article>
  );
}
