import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";

const spaces = ["AI Learning", "DSA", "Go", "System Design", "College"];

export default function SpacesPage() {
  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          eyebrow="Spaces"
          title="Topic rooms for everything you save."
          description="Spaces will group videos, documents, articles, images, and notes. For now, this page defines the frontend structure."
        />
        <Button className="w-fit" variant="outline">
          New space
        </Button>
      </div>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {spaces.map((space) => (
          <article className="rounded-lg border bg-card p-5 shadow-sm" key={space}>
            <h2 className="text-lg font-semibold">{space}</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              Placeholder collection for saved multimodal content.
            </p>
          </article>
        ))}
      </section>
    </div>
  );
}
