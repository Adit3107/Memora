import { LibraryBrowser } from "@/components/content/library-browser";
import { PageHeader } from "@/components/layout/page-header";

export default function LibraryPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Library"
        title="Everything you saved, ready to browse."
        description="Browse your saved items, filter by type, and view extracted details."
      />

      <LibraryBrowser />
    </div>
  );
}
