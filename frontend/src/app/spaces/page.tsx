import { Plus } from "lucide-react";

import { EmptyState } from "@/components/content/empty-state";
import { PageHeader } from "@/components/layout/page-header";
import { SpaceCard } from "@/components/spaces/space-card";
import { Button } from "@/components/ui/button";
import { spaces } from "@/data/content";

export default function SpacesPage() {
  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          eyebrow="Spaces"
          title="Topic rooms for everything you save."
          description="Organize videos, documents, articles, and images into focused knowledge collections. This is local mock data for Phase 1."
        />
        <Button className="w-fit cursor-not-allowed opacity-70" disabled>
          <Plus className="size-4" aria-hidden="true" />
          Create space
        </Button>
      </div>

      <section className="rounded-md border bg-card p-5 shadow-sm">
        <h2 className="text-lg font-semibold">Create Space</h2>
        <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
          Future versions will persist new Spaces. For now, this form preview
          shows the intended UI without storing data.
        </p>
        <div className="mt-5 grid gap-3 md:grid-cols-[1fr_1.4fr_auto]">
          <input
            className="h-9 rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
            disabled
            placeholder="Space name"
          />
          <input
            className="h-9 rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
            disabled
            placeholder="Description"
          />
          <Button className="cursor-not-allowed opacity-70" disabled variant="outline">
            UI only
          </Button>
        </div>
      </section>

      {spaces.length > 0 ? (
        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {spaces.map((space) => (
            <SpaceCard key={space.slug} space={space} />
          ))}
        </section>
      ) : (
        <EmptyState
          title="No spaces yet"
          description="Create your first Space to group saved videos, documents, articles, and images."
        />
      )}
    </div>
  );
}
