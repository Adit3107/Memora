import { PageHeader } from "@/components/layout/page-header";
import { SearchBrowser } from "@/components/search/search-browser";

export default function SearchPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Search"
        title="Ask for what you remember vaguely."
        description="Search across vectorized YouTube transcripts and document chunks using semantic, keyword, or hybrid retrieval."
      />

      <SearchBrowser />
    </div>
  );
}
