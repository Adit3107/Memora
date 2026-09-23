import { SaveContentForm } from "@/components/save/save-content-form";
import { PageHeader } from "@/components/layout/page-header";

export default function SavePage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Save content"
        title="Save anything into Memora."
        description="Add a YouTube URL or upload a document through the Go backend ingestion pipeline."
      />

      <SaveContentForm />
    </div>
  );
}
