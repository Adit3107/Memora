import { LibraryBrowser } from "@/components/content/library-browser";
import { PageHeader } from "@/components/layout/page-header";

export default function LibraryPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Library"
        title="Everything you saved, ready to browse."
        description="Filter mock saved content by type and open detail views. No file processing, search, or backend data is implemented in Phase 1."
      />

      <LibraryBrowser />
    </div>
  );
}
