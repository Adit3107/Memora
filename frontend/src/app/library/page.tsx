import { LibraryBrowser } from "@/components/content/library-browser";
import { PageHeader } from "@/components/layout/page-header";

export default function LibraryPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Library"
        title="Everything you saved, ready to browse."
        description="Browse backend content records, filter by type, and open detail views for ingestion metadata."
      />

      <LibraryBrowser />
    </div>
  );
}
