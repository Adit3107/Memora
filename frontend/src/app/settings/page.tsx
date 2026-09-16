import { PageHeader } from "@/components/layout/page-header";

export default function SettingsPage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Settings"
        title="Settings placeholder."
        description="This route reserves space for future preferences without implementing authentication, integrations, or backend-backed settings."
      />

      <section className="rounded-lg border bg-card p-5 shadow-sm">
        <p className="text-sm font-medium">Phase 1 only</p>
        <p className="mt-2 text-sm text-muted-foreground">
          App preferences and account controls will be added in later phases.
        </p>
      </section>
    </div>
  );
}
