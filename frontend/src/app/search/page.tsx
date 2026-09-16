import { Search } from "lucide-react";

import { PageHeader } from "@/components/layout/page-header";

export default function SearchPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Search"
        title="Ask for what you remember vaguely."
        description="Search UI is prepared for future hybrid retrieval. No semantic search, keyword search, or AI behavior is implemented in Phase 1."
      />

      <section className="rounded-lg border bg-card p-5 shadow-sm">
        <label className="text-sm font-medium" htmlFor="memora-search">
          Search Memora
        </label>
        <div className="mt-3 flex items-center gap-3 rounded-md border bg-background px-3 py-2">
          <Search className="size-4 text-muted-foreground" aria-hidden="true" />
          <input
            className="h-9 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
            id="memora-search"
            placeholder="Where did I save the vector database video?"
            type="search"
          />
        </div>
      </section>
    </div>
  );
}
