import { SaveContentForm } from "@/components/save/save-content-form";
import { PageHeader } from "@/components/layout/page-header";

export default function SavePage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Save content"
        title="Save anything into Memora."
        description="Choose a future ingestion type, assign a Space, and attach tags. This is a frontend-only workflow with no upload, extraction, embedding, or backend request."
      />

      <SaveContentForm />
    </div>
  );
}
