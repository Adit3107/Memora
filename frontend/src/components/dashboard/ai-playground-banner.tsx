"use client";

import Link from "next/link";
import {
  ArrowRight,
  Bot,
  Clock,
  FileCheck2,
  Layers,
  Sparkles,
  Zap,
} from "lucide-react";
import { motion } from "framer-motion";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const previewCapabilities = [
  {
    icon: Clock,
    title: "Video Timestamp Citations",
    description: "Deep-link directly to the exact second a concept is spoken in YouTube videos.",
  },
  {
    icon: FileCheck2,
    title: "Page-Level Document Grounding",
    description: "Answers backed by exact page numbers across PDFs, DOCX, and presentations.",
  },
  {
    icon: Layers,
    title: "Multi-Source Knowledge Synthesis",
    description: "Scope queries to your entire library, specific spaces, or select individual items.",
  },
  {
    icon: Zap,
    title: "Hallucination-Resistant RAG",
    description: "Strictly grounded in your saved pgvector embeddings with transparent source references.",
  },
];

export function AIPlaygroundBanner() {
  return (
    <motion.section
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: "easeOut" }}
      className="relative overflow-hidden rounded-2xl border border-primary/25 bg-gradient-to-br from-card via-card to-primary/5 p-6 shadow-xl shadow-primary/5 sm:p-8"
      aria-labelledby="ai-playground-banner-heading"
    >
      {/* Decorative ambient background glows */}
      <div
        className="pointer-events-none absolute -right-16 -top-16 size-72 rounded-full bg-primary/15 blur-3xl"
        aria-hidden="true"
      />
      <div
        className="pointer-events-none absolute -bottom-16 -left-16 size-60 rounded-full bg-cyan-500/10 blur-3xl"
        aria-hidden="true"
      />

      <div className="relative z-10 flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-2xl space-y-3">
          {/* Status Badges */}
          <div className="flex flex-wrap items-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-full border border-primary/40 bg-primary/10 px-3 py-1 text-xs font-semibold uppercase tracking-wider text-primary">
              <span className="size-2 animate-pulse rounded-full bg-primary" />
              Launching Soon
            </span>
            <span className="inline-flex items-center gap-1 rounded-full border border-border/80 bg-background/80 px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
              <Sparkles className="size-3 text-primary" />
              AI Playground & Grounded Synthesis
            </span>
          </div>

          <h2
            id="ai-playground-banner-heading"
            className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl"
          >
            Chat with your second memory is arriving soon.
          </h2>

          <p className="text-sm leading-relaxed text-muted-foreground sm:text-base">
            We are fine-tuning our conversational RAG engine. Soon you will be
            able to ask open-ended questions across all your saved YouTube
            transcripts, documents, and notes—receiving syntheses with
            verifiable video timestamps and page numbers.
          </p>
        </div>

        {/* Action Button */}
        <div className="flex shrink-0 flex-wrap items-center gap-3">
          <Link
            href="/app/search"
            className={cn(
              buttonVariants({ variant: "outline" }),
              "gap-2 border-primary/30 bg-background/60 backdrop-blur transition-all hover:border-primary/60 hover:bg-accent"
            )}
          >
            <span>Explore Search in the meantime</span>
            <ArrowRight className="size-4" />
          </Link>
          <Link
            href="/app/ai"
            className={cn(
              buttonVariants({ variant: "default" }),
              "gap-2 shadow-md shadow-primary/20"
            )}
          >
            <Bot className="size-4" />
            <span>Preview Feature</span>
          </Link>
        </div>
      </div>

      {/* Sneak-Peek Feature Cards Grid */}
      <div className="relative z-10 mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {previewCapabilities.map((cap) => {
          const Icon = cap.icon;
          return (
            <div
              key={cap.title}
              className="group flex flex-col justify-between rounded-xl border border-border/60 bg-background/60 p-4 backdrop-blur transition-colors hover:border-primary/40 hover:bg-card/80"
            >
              <div className="space-y-2">
                <div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary transition-transform group-hover:scale-105">
                  <Icon className="size-4" />
                </div>
                <h3 className="text-xs font-semibold text-foreground">
                  {cap.title}
                </h3>
                <p className="text-xs leading-5 text-muted-foreground">
                  {cap.description}
                </p>
              </div>
            </div>
          );
        })}
      </div>
    </motion.section>
  );
}
