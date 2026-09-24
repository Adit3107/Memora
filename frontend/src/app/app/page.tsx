import Link from "next/link";

import { AIPlaygroundBanner } from "@/components/dashboard/ai-playground-banner";
import { ContentTypeCard } from "@/components/dashboard/content-type-card";
import { LiveDashboard } from "@/components/dashboard/live-dashboard";
import { QuickActionCard } from "@/components/dashboard/quick-action-card";
import { SectionHeading } from "@/components/dashboard/section-heading";
import { SpaceSummaryCard } from "@/components/dashboard/space-summary-card";
import { buttonVariants } from "@/components/ui/button";
import {
  contentTypeSummaries,
  quickActions,
  spaceSummaries,
} from "@/data/dashboard";
import { cn } from "@/lib/utils";

export default function AppHomePage() {
  return (
    <div className="space-y-8">
      <section className="rounded-md border bg-card p-5 shadow-sm lg:p-6">
        <div className="grid gap-6 lg:grid-cols-[1fr_auto] lg:items-start">
          <div className="max-w-3xl space-y-3">
            <p className="text-sm font-medium uppercase tracking-normal text-muted-foreground">
              Home
            </p>
            <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl">
              Good evening, Aditya.
            </h1>
            <p className="text-base leading-7 text-muted-foreground">
              Your memory, organized. Save videos and documents, let Memora
              understand them, then search everything when you need it.
            </p>
          </div>

          <div className="flex flex-wrap gap-2">
            <Link
              className={cn(buttonVariants({ variant: "default" }), "w-fit")}
              href="/app/save"
            >
              Add memory
            </Link>
            <Link
              className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
              href="/app/search"
            >
              Search memory
            </Link>
          </div>
        </div>
      </section>

      <Link
        className="flex h-14 items-center justify-between rounded-md border bg-card px-4 text-left shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
        href="/app/search"
      >
        <span className="text-sm text-muted-foreground">Search your memory...</span>
        <kbd className="rounded-md border bg-background px-2 py-1 font-mono text-xs text-muted-foreground">
          Ctrl K
        </kbd>
      </Link>

      {/* AI Playground Feature on Hold - Attractive Theme-Relevant Launching Soon Banner */}
      <AIPlaygroundBanner />

      <LiveDashboard />

      <section className="space-y-4" aria-labelledby="quick-actions-heading">
        <SectionHeading
          id="quick-actions-heading"
          title="Quick actions"
          description="Capture, browse, and retrieve knowledge from the live backend."
        />
        <div className="grid gap-4 md:grid-cols-3">
          {quickActions.map((action) => (
            <QuickActionCard
              action={{ ...action, href: `/app${action.href}` }}
              key={action.title}
            />
          ))}
        </div>
      </section>

      <div className="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
        <section className="space-y-4" aria-labelledby="spaces-heading">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <SectionHeading
              id="spaces-heading"
              title="Spaces overview"
              description="Topic collections for mixed saved content."
            />
            <Link
              className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
              href="/app/spaces"
            >
              View spaces
            </Link>
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            {spaceSummaries.map((space) => (
              <SpaceSummaryCard key={space.name} space={space} />
            ))}
          </div>
        </section>

        <section className="space-y-4" aria-labelledby="content-types-heading">
          <SectionHeading
            id="content-types-heading"
            title="Content-type overview"
            description="Formats Memora can capture, process, and retrieve."
          />
          <div className="grid gap-3 sm:grid-cols-2">
            {contentTypeSummaries.map((item) => (
              <ContentTypeCard item={item} key={item.type} />
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
