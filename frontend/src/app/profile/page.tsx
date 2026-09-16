import { PageHeader } from "@/components/layout/page-header";

export default function ProfilePage() {
  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Profile"
        title="User area placeholder."
        description="Authentication and account data belong to later phases. This page reserves the user/profile surface for the app shell."
      />

      <section className="rounded-lg border bg-card p-5 shadow-sm">
        <p className="text-sm font-medium">Memora User</p>
        <p className="mt-2 text-sm text-muted-foreground">
          Profile details will appear here after authentication is introduced.
        </p>
      </section>
    </div>
  );
}
