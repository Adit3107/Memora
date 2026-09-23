import { ThemeToggle } from "@/components/theme/theme-toggle";
import type { ReactNode } from "react";

export default function SettingsPage() {
  return (
    <div className="space-y-8">
      <div>
        <p className="text-sm font-medium uppercase tracking-normal text-muted-foreground">
          Settings
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-normal">
          Tune your Memora workspace.
        </h1>
      </div>

      <section className="grid gap-4 lg:grid-cols-2">
        <SettingsCard title="Account">
          <div className="flex items-center gap-4">
            <div className="flex size-12 items-center justify-center rounded-full bg-primary text-lg font-semibold text-primary-foreground">
              A
            </div>
            <div>
              <p className="font-medium">Aditya</p>
              <p className="text-sm text-muted-foreground">demo@memora.local</p>
            </div>
          </div>
        </SettingsCard>

        <SettingsCard title="Appearance">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium">Theme</p>
              <p className="text-sm text-muted-foreground">
                Cycle light, dark, and system preference.
              </p>
            </div>
            <ThemeToggle />
          </div>
        </SettingsCard>

        <SettingsCard title="Preferences">
          <div className="grid gap-3 sm:grid-cols-2">
            <label className="space-y-2 text-sm font-medium">
              Default library view
              <select className="h-10 w-full rounded-md border bg-background px-3 text-sm">
                <option>Recently added</option>
                <option>A-Z</option>
              </select>
            </label>
            <label className="space-y-2 text-sm font-medium">
              Search mode
              <select className="h-10 w-full rounded-md border bg-background px-3 text-sm">
                <option>Hybrid</option>
                <option>Semantic</option>
                <option>Keyword</option>
              </select>
            </label>
          </div>
        </SettingsCard>

        <SettingsCard title="Storage">
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Demo workspace</span>
              <span className="font-medium">Backend managed</span>
            </div>
            <div className="h-2 overflow-hidden rounded-full bg-muted">
              <div className="h-full w-2/5 rounded-full bg-primary" />
            </div>
          </div>
        </SettingsCard>

        <SettingsCard title="About">
          <p className="text-sm leading-6 text-muted-foreground">
            Memora frontend. Phase 6-ready ingestion and retrieval UI.
          </p>
        </SettingsCard>
      </section>
    </div>
  );
}

function SettingsCard({
  children,
  title,
}: {
  children: ReactNode;
  title: string;
}) {
  return (
    <article className="rounded-lg border bg-card p-5 shadow-sm">
      <h2 className="mb-4 text-lg font-semibold">{title}</h2>
      {children}
    </article>
  );
}
