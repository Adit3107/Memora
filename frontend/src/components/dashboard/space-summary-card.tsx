import type { SpaceSummary } from "@/types/dashboard";

type SpaceSummaryCardProps = {
  space: SpaceSummary;
};

export function SpaceSummaryCard({ space }: SpaceSummaryCardProps) {
  return (
    <article className="rounded-md border bg-card p-4 shadow-sm">
      <h3 className="text-base font-semibold">{space.name}</h3>
      <p className="mt-2 text-sm text-muted-foreground">
        {space.itemCount} saved items
      </p>
      <p className="mt-1 text-xs text-muted-foreground">{space.lastUpdated}</p>
    </article>
  );
}
