import Link from "next/link";

import { ContentTypeCard } from "@/components/dashboard/content-type-card";
import { OverviewCard } from "@/components/dashboard/overview-card";
import { QuickActionCard } from "@/components/dashboard/quick-action-card";
import { RecentContentCard } from "@/components/dashboard/recent-content-card";
import { SectionHeading } from "@/components/dashboard/section-heading";
import { SpaceSummaryCard } from "@/components/dashboard/space-summary-card";
import { buttonVariants } from "@/components/ui/button";
import {
  contentTypeSummaries,
  dashboardOverview,
  quickActions,
  recentContent,
  spaceSummaries,
} from "@/data/dashboard";
import { cn } from "@/lib/utils";

export default function DashboardPage() {
  return (
    <div className="space-y-8">
      <section className="rounded-md border bg-card p-5 shadow-sm lg:p-6">
        <div className="grid gap-6 lg:grid-cols-[1fr_auto] lg:items-start">
          <div className="max-w-3xl space-y-3">
            <p className="text-sm font-medium uppercase tracking-normal text-muted-foreground">
              Dashboard
            </p>
            <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl">
              Your second memory for everything worth saving.
            </h1>
            <p className="text-base leading-7 text-muted-foreground">
              Memora turns videos, documents, articles, and images into an
              organized personal knowledge base. This dashboard is static for
              Phase 1, ready for backend, ingestion, search, and AI integration
              later.
            </p>
          </div>

          <Link
            className={cn(buttonVariants({ variant: "default" }), "w-fit")}
            href="/search"
          >
            Search memory
          </Link>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-3" aria-label="Overview">
        {dashboardOverview.map((item) => (
          <OverviewCard item={item} key={item.label} />
        ))}
      </section>

      <section className="space-y-4" aria-labelledby="quick-actions-heading">
        <SectionHeading
          id="quick-actions-heading"
          title="Quick actions"
          description="Action surfaces are prepared now; saving and uploading remain Phase 1 UI placeholders."
        />
        <div className="grid gap-4 md:grid-cols-3">
          {quickActions.map((action) => (
            <QuickActionCard action={action} key={action.title} />
          ))}
        </div>
      </section>

      <section className="space-y-4" aria-labelledby="recent-content-heading">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <SectionHeading
            id="recent-content-heading"
            title="Recent content"
            description="Mock examples showing how saved items will communicate type, source, metadata, space, and tags."
          />
          <Link
            className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
            href="/library"
          >
            View library
          </Link>
        </div>
        <div className="grid gap-3">
          {recentContent.map((item) => (
            <RecentContentCard item={item} key={item.title} />
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
              href="/spaces"
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
            description="A quick scan of the formats Memora will understand."
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
