"use client";

import Link from "next/link";
import {
  ArrowLeft,
  Bot,
  CheckCircle2,
  Clock,
  ExternalLink,
  FileSearch,
  Layers,
  Search,
  Sparkles,
  Video,
} from "lucide-react";
import { motion } from "framer-motion";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// =============================================================================
// PAUSED FEATURE: AI Playground Page
// The underlying RAG and conversational engine is temporarily on hold.
// In the meantime, this view displays the preview roadmap and capabilities.
// The original playground component is isolated in:
//   src/paused_features/rag/ai-playground.tsx (commented out)
// =============================================================================

const roadmapHighlights = [
  {
    icon: Video,
    title: "YouTube Timestamp Jump",
    badge: "Vector Grounding",
    description:
      "When asking questions about technical video tutorials, answers will include click-to-play timestamps pointing to the exact second where code or theory was explained.",
  },
  {
    icon: FileSearch,
    title: "Document Page Citations",
    badge: "Multi-Format",
    description:
      "Every claim synthesized from PDFs, DOCX files, or lecture slides cites the specific page number with extracted context previews.",
  },
  {
    icon: Layers,
    title: "Configurable Synthesis Scopes",
    badge: "Custom Context",
    description:
      "Toggle between querying your entire knowledge base, an isolated Space (e.g. 'System Design'), or individually selected sources.",
  },
  {
    icon: Clock,
    title: "Hallucination Safeguards",
    badge: "Privacy First",
    description:
      "Responses are strictly grounded in your PostgreSQL + pgvector stored chunks. When information is missing, Memora transparently informs you rather than guessing.",
  },
];

export default function AIPlaygroundPage() {
  return (
    <div className="mx-auto max-w-5xl space-y-8 py-4">
      {/* Top Breadcrumb & Status */}
      <div className="flex items-center justify-between">
        <Link
          href="/app"
          className="inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          <span>Back to Dashboard</span>
        </Link>
        <span className="inline-flex items-center gap-1.5 rounded-full border border-primary/40 bg-primary/10 px-3 py-1 text-xs font-semibold uppercase tracking-wider text-primary">
          <span className="size-2 animate-pulse rounded-full bg-primary" />
          On Hold • Launching Soon
        </span>
      </div>

      {/* Hero Showcase Card */}
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
        className="relative overflow-hidden rounded-3xl border border-primary/25 bg-gradient-to-b from-card via-card to-primary/5 p-8 shadow-2xl sm:p-12"
      >
        <div
          className="pointer-events-none absolute -right-20 -top-20 size-80 rounded-full bg-primary/20 blur-3xl"
          aria-hidden="true"
        />
        <div
          className="pointer-events-none absolute -bottom-20 -left-20 size-80 rounded-full bg-cyan-500/10 blur-3xl"
          aria-hidden="true"
        />

        <div className="relative z-10 max-w-3xl space-y-6">
          <div className="inline-flex size-14 items-center justify-center rounded-2xl border border-primary/40 bg-primary/15 text-primary shadow-lg shadow-primary/20">
            <Bot className="size-7" />
          </div>

          <div className="space-y-3">
            <h1 className="text-3xl font-extrabold tracking-tight text-foreground sm:text-5xl">
              Memora AI Playground
            </h1>
            <p className="text-base font-normal leading-relaxed text-muted-foreground sm:text-lg">
              We are currently finalizing the grounded RAG architecture and
              citations engine. Soon, you will be able to have interactive,
              multi-turn conversations with all your saved videos, documents,
              articles, and notes.
            </p>
          </div>

          {/* Key Checklist */}
          <div className="grid gap-2.5 pt-2 sm:grid-cols-2">
            {[
              "PostgreSQL + pgvector chunk retrieval ready",
              "SentenceTransformers 384-dim embeddings active",
              "Hybrid reciprocal rank fusion in place",
              "Grounded citation verification in progress",
            ].map((item) => (
              <div
                key={item}
                className="flex items-center gap-2.5 text-xs font-medium text-foreground sm:text-sm"
              >
                <CheckCircle2 className="size-4 shrink-0 text-emerald-400" />
                <span>{item}</span>
              </div>
            ))}
          </div>

          <div className="flex flex-wrap items-center gap-3 pt-4">
            <Link
              href="/app/search"
              className={cn(
                buttonVariants({ variant: "default" }),
                "gap-2 shadow-lg shadow-primary/20"
              )}
            >
              <Search className="size-4" />
              <span>Use Hybrid Search Now</span>
            </Link>
            <Link
              href="/app/library"
              className={cn(
                buttonVariants({ variant: "outline" }),
                "gap-2 border-border/80 bg-background/60"
              )}
            >
              <span>Browse Saved Knowledge</span>
              <ExternalLink className="size-3.5" />
            </Link>
          </div>
        </div>
      </motion.div>

      {/* Upcoming Capabilities Grid */}
      <section className="space-y-4" aria-labelledby="roadmap-heading">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary">
            <Sparkles className="size-3.5" />
            <span>Architecture Preview</span>
          </div>
          <h2
            id="roadmap-heading"
            className="text-xl font-bold tracking-tight text-foreground sm:text-2xl"
          >
            What to expect when AI Playground goes live
          </h2>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {roadmapHighlights.map((feature) => {
            const Icon = feature.icon;
            return (
              <div
                key={feature.title}
                className="flex flex-col justify-between rounded-2xl border border-border/70 bg-card p-6 shadow-sm transition-all hover:border-primary/40 hover:shadow-md"
              >
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
                      <Icon className="size-5" />
                    </div>
                    <span className="rounded-md border border-border bg-background px-2.5 py-0.5 text-[11px] font-medium text-muted-foreground">
                      {feature.badge}
                    </span>
                  </div>
                  <div>
                    <h3 className="text-base font-semibold text-foreground">
                      {feature.title}
                    </h3>
                    <p className="mt-1 text-sm leading-6 text-muted-foreground">
                      {feature.description}
                    </p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}
