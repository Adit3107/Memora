import { SaveContentForm } from "@/components/save/save-content-form";
import { PageHeader } from "@/components/layout/page-header";

export default function SavePage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Save content"
        title="Save anything into Mindshelf."
        description="Add an Instagram Reel, Facebook Reel, YouTube URL, or document into your intelligent shelf."
      />

      <SaveContentForm />
    </div>
  );
}
