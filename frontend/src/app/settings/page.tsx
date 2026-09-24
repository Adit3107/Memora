"use client";

import { useUser } from "@clerk/nextjs";
import Link from "next/link";
import { ArrowRight, User } from "lucide-react";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import type { ReactNode } from "react";

export default function SettingsPage() {
  const { user, isLoaded } = useUser();
  const displayName = user?.fullName || user?.firstName || "Mindshelf User";
  const displayEmail = user?.primaryEmailAddress?.emailAddress || "user@mindshelf.app";
  const initial = user?.firstName?.[0] || user?.fullName?.[0] || "M";

  return (
    <div className="space-y-8">
      <div>
        <p className="text-sm font-medium uppercase tracking-normal text-muted-foreground">
          Settings
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-normal">
          Tune your Mindshelf workspace.
        </h1>
      </div>

      <section className="grid gap-4 lg:grid-cols-2">
        <SettingsCard title="Account">
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-4">
              {user?.imageUrl ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  alt=""
                  className="size-12 rounded-full border border-border object-cover"
                  src={user.imageUrl}
                />
              ) : (
                <div className="flex size-12 items-center justify-center rounded-full bg-primary text-lg font-semibold text-primary-foreground">
                  {initial}
                </div>
              )}
              <div className="min-w-0 flex-1">
                <p className="font-medium truncate">{displayName}</p>
                <p className="text-sm text-muted-foreground truncate">{displayEmail}</p>
              </div>
            </div>

            <Link
              className="inline-flex items-center gap-2 rounded-lg border border-border/80 bg-background/60 px-3.5 py-2 text-xs font-medium text-foreground transition hover:bg-accent hover:text-accent-foreground"
              href="/app/profile"
            >
              <User className="size-3.5 text-primary" />
              <span>Manage full account & security profile</span>
              <ArrowRight className="ml-auto size-3 text-muted-foreground" />
            </Link>
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
            Mindshelf frontend. Intelligent video and document second memory.
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
