import { PageHeader } from "@/components/layout/page-header";

const libraryItems = [
  { title: "Vector Database Short", type: "Video", location: "AI Learning" },
  { title: "RAG Notes.pdf", type: "Document", location: "System Design" },
  { title: "Go Pipelines Article", type: "Article", location: "Go" },
  { title: "DSA Revision Sheet", type: "Spreadsheet", location: "DSA" },
];

export default function LibraryPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Library"
        title="One library for every source type."
        description="This view will become the unified content library. Current entries are static UI examples only."
      />

      <section className="overflow-hidden rounded-lg border bg-card shadow-sm">
        <div className="grid grid-cols-[1.4fr_0.6fr_0.8fr] border-b px-4 py-3 text-sm font-medium text-muted-foreground">
          <span>Title</span>
          <span>Type</span>
          <span>Space</span>
        </div>
        {libraryItems.map((item) => (
          <div
            className="grid grid-cols-[1.4fr_0.6fr_0.8fr] border-b px-4 py-4 text-sm last:border-b-0"
            key={item.title}
          >
            <span className="font-medium">{item.title}</span>
            <span className="text-muted-foreground">{item.type}</span>
            <span className="text-muted-foreground">{item.location}</span>
          </div>
        ))}
      </section>
    </div>
  );
}
