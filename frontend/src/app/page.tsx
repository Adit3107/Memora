import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";

const highlights = [
  { label: "Saved items", value: "24", detail: "Across videos and docs" },
  { label: "Spaces", value: "5", detail: "AI Learning, DSA, Go" },
  { label: "Recent searches", value: "8", detail: "Ready for Phase 6" },
];

export default function DashboardPage() {
  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          eyebrow="Dashboard"
          title="Your saved knowledge, organized for recall."
          description="This dashboard establishes the main Memora workspace. The content here is static until backend, ingestion, search, and AI phases are built."
        />
        <Button className="w-fit" variant="outline">
          Save content
        </Button>
      </div>

      <section className="grid gap-4 md:grid-cols-3" aria-label="Overview">
        {highlights.map((item) => (
          <div className="rounded-lg border bg-card p-5 shadow-sm" key={item.label}>
            <p className="text-sm text-muted-foreground">{item.label}</p>
            <p className="mt-3 text-3xl font-semibold">{item.value}</p>
            <p className="mt-2 text-sm text-muted-foreground">{item.detail}</p>
          </div>
        ))}
      </section>

      <section className="rounded-lg border bg-card p-5 shadow-sm">
        <div className="border-b pb-4">
          <h2 className="text-lg font-semibold">Recent activity</h2>
          <p className="text-sm text-muted-foreground">
            Static examples for the Phase 1 interface.
          </p>
        </div>
        <div className="divide-y">
          {[
            "Vector database explainer saved to AI Learning",
            "RAG notes added to System Design",
            "Go concurrency article added to Library",
          ].map((activity) => (
            <div className="py-4 text-sm" key={activity}>
              {activity}
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
