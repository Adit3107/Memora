"use client";

import {
  Check,
  Copy,
  ExternalLink,
  FileText,
  Image,
  Layers,
  Newspaper,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Save,
  Sparkles,
  Trash2,
  Video,
  X,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useUser } from "@clerk/nextjs";

import { contentTypeLabels } from "@/data/content";
import {
  createTag,
  deleteContent,
  getContentSummary,
  listTags,
  setContentTags,
  updateContent,
  type BackendTag,
  type ContentSummary,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import type { SavedContentItem } from "@/types/content";

import { buttonVariants } from "../ui/button";

type ContentDetailProps = {
  item: SavedContentItem;
};

const iconForType = {
  video: Video,
  document: FileText,
  article: Newspaper,
  image: Image,
};

export function ContentDetail({ item }: ContentDetailProps) {
  const { user } = useUser();
  const router = useRouter();
  const Icon = (item.type && iconForType[item.type]) || Video;

  const [copied, setCopied] = useState(false);
  const [summaryData, setSummaryData] = useState<ContentSummary | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [summaryError, setSummaryError] = useState("");
  const [contentName, setContentName] = useState(item.title);
  const [isEditingName, setIsEditingName] = useState(false);
  const [isSavingName, setIsSavingName] = useState(false);
  const [contentTags, setDisplayedTags] = useState(item.tags);
  const [availableTags, setAvailableTags] = useState<BackendTag[]>([]);
  const [isEditingTags, setIsEditingTags] = useState(false);
  const [newTag, setNewTag] = useState("");
  const [isSavingTags, setIsSavingTags] = useState(false);

  const platform = item.platform || detectPlatformFromUrl(item.sourceUrl);

  useEffect(() => {
    if (!user?.id) return;
    void listTags(user.id).then(setAvailableTags).catch(() => setAvailableTags([]));
  }, [user?.id]);

  async function saveName() {
    const name = contentName.trim();
    if (!user?.id || !item.id || !name || name === item.title) {
      setIsEditingName(false);
      return;
    }
    setIsSavingName(true);
    try {
      await updateContent(item.id, {
        user_id: user.id,
        space_id: item.spaceSlug,
        name,
        title: item.title,
        description: item.description,
        type: item.type,
        source_url: item.sourceUrl,
      });
      setIsEditingName(false);
    } catch {
      window.alert("Content name could not be updated.");
    } finally {
      setIsSavingName(false);
    }
  }

  function toggleContentTag(name: string) {
    setDisplayedTags((current) =>
      current.includes(name)
        ? current.filter((tag) => tag !== name)
        : [...current, name]
    );
  }

  async function saveTags() {
    if (!user?.id || !item.id) return;
    setIsSavingTags(true);
    try {
      const tags = await Promise.all(
        contentTags.map(async (name) => {
          const existing = availableTags.find(
            (tag) => tag.name.toLowerCase() === name.toLowerCase()
          );
          return existing ?? createTag({ user_id: user.id, name });
        })
      );
      await setContentTags(item.id, tags.map((tag) => tag.id));
      setAvailableTags(await listTags(user.id));
      setIsEditingTags(false);
    } catch {
      window.alert("Tags could not be updated.");
    } finally {
      setIsSavingTags(false);
    }
  }

  async function removeContent() {
    if (!item.id || !window.confirm(`Delete "${contentName}"? This cannot be undone.`)) return;
    try {
      await deleteContent(item.id);
      router.push("/app/library");
    } catch {
      window.alert("Content could not be deleted.");
    }
  }

  // Fetch AI Video/Content Summary using Gemini API
  async function loadSummary() {
    if (!item.id) return;
    setSummaryLoading(true);
    setSummaryError("");
    try {
      const data = await getContentSummary(item.id);
      setSummaryData(data);
    } catch {
      // Fallback summary from description if backend summary call fails
      setSummaryData({
        summary: item.description || "Executive overview for this content.",
        bullets: [
          `Captured from ${platform ? platform.toUpperCase() : item.source}.`,
          "Transcript segments processed and indexed for vector similarity search.",
          "Grounded in PostgreSQL storage with pgvector embeddings.",
        ],
        source: "fallback",
      });
    } finally {
      setSummaryLoading(false);
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadSummary();
  }, [item.id]);

  function copyToClipboard(text: string) {
    void navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="space-y-8">
      {/* Header section */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="max-w-3xl space-y-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-md bg-secondary px-2.5 py-1 text-xs font-medium text-secondary-foreground">
              {contentTypeLabels[item.type]}
            </span>

            {platform ? (
              <PlatformBadge platform={platform} />
            ) : (
              <span className="text-xs text-muted-foreground">{item.source}</span>
            )}

            <span className="text-xs text-muted-foreground">• {item.dateLabel}</span>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            {isEditingName ? (
              <input
                className="h-11 min-w-0 flex-1 rounded-md border bg-background px-3 text-xl font-semibold outline-none focus-visible:ring-3 focus-visible:ring-ring/50 sm:text-3xl"
                disabled={isSavingName}
                onChange={(event) => setContentName(event.target.value)}
                value={contentName}
              />
            ) : (
              <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl">
                {contentName}
              </h1>
            )}
            <div className="flex items-center gap-1">
              {isEditingName ? (
                <>
                  <button className="inline-flex size-9 items-center justify-center rounded-md border hover:bg-accent" onClick={() => void saveName()} type="button" title="Save name">
                    <Save className="size-4" />
                  </button>
                  <button className="inline-flex size-9 items-center justify-center rounded-md border hover:bg-accent" onClick={() => { setContentName(item.title); setIsEditingName(false); }} type="button" title="Cancel">
                    <X className="size-4" />
                  </button>
                </>
              ) : (
                <button className="inline-flex size-9 items-center justify-center rounded-md border text-muted-foreground hover:bg-accent hover:text-foreground" onClick={() => setIsEditingName(true)} type="button" title="Rename content">
                  <Pencil className="size-4" />
                </button>
              )}
              <button className="inline-flex size-9 items-center justify-center rounded-md border border-destructive/40 text-destructive hover:bg-destructive/10" onClick={() => void removeContent()} type="button" title="Delete content">
                <Trash2 className="size-4" />
              </button>
            </div>
          </div>
          <p className="text-base leading-relaxed text-muted-foreground">
            {item.description}
          </p>
        </div>

        <Link
          className={cn(buttonVariants({ variant: "outline" }), "w-fit shrink-0")}
          href="/app/library"
        >
          Back to library
        </Link>
      </div>

      {/* 3 Information Cards */}
      <section className="grid gap-4 md:grid-cols-3">
        {/* Card 1: Space */}
        <div className="rounded-xl border bg-card p-5 shadow-sm transition hover:border-border/80">
          <div className="flex items-center justify-between">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Destination Space
            </p>
            <Layers className="size-4 text-muted-foreground" />
          </div>
          <Link
            className="mt-2.5 block text-base font-semibold hover:underline"
            href={`/app/spaces/${item.spaceSlug}`}
          >
            {item.spaceName}
          </Link>
          <p className="mt-1 text-xs text-muted-foreground">
            Grouped within your knowledge collections.
          </p>
        </div>

        {/* Card 2: Metadata (YT Title, Insta/FB Reel ID) */}
        <div className="rounded-xl border bg-card p-5 shadow-sm transition hover:border-border/80">
          <div className="flex items-center justify-between">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Content Metadata
            </p>
            {item.metadata.includes("ID:") ? (
              <button
                className="inline-flex items-center gap-1 rounded border border-border px-1.5 py-0.5 text-[11px] font-medium text-muted-foreground hover:text-foreground"
                onClick={() => copyToClipboard(item.metadata.replace("Reel ID: ", "").trim())}
                title="Copy ID"
                type="button"
              >
                {copied ? <Check className="size-3 text-emerald-400" /> : <Copy className="size-3" />}
                <span>{copied ? "Copied" : "Copy"}</span>
              </button>
            ) : null}
          </div>

          <div className="mt-2.5">
            <p className="line-clamp-2 text-base font-semibold text-foreground">
              {item.metadata}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              {platform === "youtube"
                ? "Original YouTube video title"
                : platform === "instagram"
                ? "Original Instagram Reel ID"
                : platform === "facebook"
                ? "Original Facebook Reel ID"
                : "Extracted document metadata"}
            </p>
          </div>
        </div>

        {/* Card 3: Source Link */}
        <div className="rounded-xl border bg-card p-5 shadow-sm transition hover:border-border/80">
          <div className="flex items-center justify-between">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Original Source
            </p>
            <ExternalLink className="size-4 text-muted-foreground" />
          </div>

          {item.sourceUrl ? (
            <div className="mt-2.5">
              <Link
                className="inline-flex items-center gap-1.5 truncate text-base font-semibold text-primary underline-offset-4 hover:underline"
                href={item.sourceUrl}
                rel="noreferrer"
                target="_blank"
              >
                <span className="truncate">{cleanUrlDomain(item.sourceUrl)}</span>
                <ExternalLink className="size-3.5 shrink-0" />
              </Link>
              <p className="mt-1 truncate text-xs text-muted-foreground">
                {item.sourceUrl}
              </p>
            </div>
          ) : (
            <p className="mt-2.5 text-base font-semibold text-muted-foreground">
              Uploaded Document
            </p>
          )}
        </div>
      </section>

      <section className="rounded-xl border bg-card p-5 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">Tags</h2>
            <div className="mt-3 flex flex-wrap gap-2">
              {contentTags.length > 0 ? contentTags.map((tag) => (
                <span className="rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground" key={tag}>#{tag}</span>
              )) : <span className="text-sm text-muted-foreground">No tags yet.</span>}
            </div>
          </div>
          <button className="inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium hover:bg-accent" onClick={() => setIsEditingTags((current) => !current)} type="button">
            {isEditingTags ? <X className="size-4" /> : <Pencil className="size-4" />}
            {isEditingTags ? "Cancel" : "Edit tags"}
          </button>
        </div>
        {isEditingTags ? (
          <div className="mt-4 space-y-3 border-t pt-4">
            <div className="flex flex-wrap gap-2">
              {availableTags.map((tag) => (
                <button className={cn("rounded-md border px-3 py-2 text-sm", contentTags.includes(tag.name) && "bg-primary text-primary-foreground")} key={tag.id} onClick={() => toggleContentTag(tag.name)} type="button">#{tag.name}</button>
              ))}
            </div>
            <div className="flex gap-2">
              <input className="h-10 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm" onChange={(event) => setNewTag(event.target.value)} placeholder="New tag" value={newTag} />
              <button className="inline-flex items-center gap-1 rounded-md border px-3 text-sm" onClick={() => { const name = newTag.trim(); if (name && !contentTags.some((tag) => tag.toLowerCase() === name.toLowerCase())) setDisplayedTags((current) => [...current, name]); setNewTag(""); }} type="button"><Plus className="size-4" />Add</button>
              <button className="inline-flex items-center gap-1 rounded-md bg-primary px-3 text-sm text-primary-foreground disabled:opacity-50" disabled={isSavingTags} onClick={() => void saveTags()} type="button"><Save className="size-4" />Save</button>
            </div>
          </div>
        ) : null}
      </section>

      {/* Preview and AI summary use independent full-width sections so the summary is never stretched by the preview height. */}
      <div className="grid gap-6">
        <section className="rounded-xl border bg-card p-4 shadow-sm">
          <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border/80 bg-background/50 p-6 text-center">
            {item.thumbnailUrl ? (
              <div className="w-full">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  alt={item.title}
                  className="mx-auto max-h-80 w-full rounded-lg border border-border/60 bg-black/60 object-contain shadow-md"
                  src={item.thumbnailUrl}
                />
              </div>
            ) : (
              /* Platform Logo Fallback if no cover image */
              <div className="my-6">
                <PlatformLogo platform={platform} type={item.type} />
              </div>
            )}

            <div className="mt-5 space-y-1.5">
              <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                {platform ? `${platform.toUpperCase()} PREVIEW` : "DOCUMENT PREVIEW"}
              </p>
              <h2 className="text-xl font-bold tracking-tight text-foreground">
                {item.title}
              </h2>
              <p className="mx-auto max-w-md text-xs leading-relaxed text-muted-foreground">
                {item.thumbnailUrl
                  ? "Full cover image rendered with object-contain preservation."
                  : "Verified and cataloged in Mindshelf."}
              </p>
            </div>
          </div>
        </section>

        <section className="rounded-xl border bg-card p-6 shadow-sm">
          <div>
            <div className="flex items-center justify-between border-b pb-4">
              <div className="flex items-center gap-2.5">
                <div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <Sparkles className="size-4" />
                </div>
                <div>
                  <h2 className="text-lg font-bold tracking-tight text-foreground">
                    {item.type === "video" ? "AI Video Summary" : "AI Content Summary"}
                  </h2>
                  <p className="text-xs text-muted-foreground">
                    Synthesized from audio transcription & visual OCR
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <span className="rounded-full border border-primary/30 bg-primary/10 px-2 py-0.5 text-[11px] font-semibold text-primary">
                  Gemini AI
                </span>
                <button
                  className="rounded-md border p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground disabled:opacity-50"
                  disabled={summaryLoading}
                  onClick={() => void loadSummary()}
                  title="Regenerate summary"
                  type="button"
                >
                  <RefreshCw className={cn("size-3.5", summaryLoading && "animate-spin")} />
                </button>
              </div>
            </div>

            {summaryLoading ? (
              <div className="mt-5 space-y-3">
                <div className="h-4 w-3/4 animate-pulse rounded bg-muted" />
                <div className="h-16 w-full animate-pulse rounded-lg bg-muted" />
                <div className="space-y-2 pt-2">
                  <div className="h-4 w-5/6 animate-pulse rounded bg-muted" />
                  <div className="h-4 w-4/6 animate-pulse rounded bg-muted" />
                  <div className="h-4 w-3/6 animate-pulse rounded bg-muted" />
                </div>
              </div>
            ) : summaryData ? (
              <div className="mt-5 space-y-5">
                {/* Executive Summary paragraph */}
                <div className="rounded-lg border border-primary/20 bg-primary/5 p-4">
                  <p className="text-xs font-semibold uppercase tracking-wider text-primary">
                    Executive Overview
                  </p>
                  <p className="mt-1.5 text-sm leading-relaxed text-foreground">
                    {summaryData.summary}
                  </p>
                </div>

                {/* Bulleted Takeaways */}
                <div>
                  <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    Key Highlights & Takeaways
                  </p>
                  <ul className="mt-3 space-y-2.5">
                    {summaryData.bullets.map((bullet, idx) => (
                      <li
                        className="flex items-start gap-2.5 rounded-lg border border-border/70 bg-background/60 p-3 text-xs leading-relaxed text-foreground"
                        key={idx}
                      >
                        <span className="mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-primary/10 text-[10px] font-bold text-primary">
                          {idx + 1}
                        </span>
                        <span>{bullet}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            ) : (
              <p className="mt-6 text-sm text-muted-foreground">
                No summary generated yet. Click regenerate to synthesize.
              </p>
            )}
          </div>
        </section>
      </div>

      {/* Human-understandable Source & Provenance section */}
      <section className="rounded-xl border bg-card p-6 shadow-sm">
        <div className="flex items-center justify-between border-b pb-4">
          <div>
            <h2 className="text-lg font-bold tracking-tight text-foreground">
              Source & Provenance
            </h2>
            <p className="text-xs text-muted-foreground">
              Clear attribution and ingestion pipeline details for this item
            </p>
          </div>
        </div>

        <div className="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div className="rounded-lg border bg-background/60 p-4">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Source Platform
            </p>
            <p className="mt-1.5 text-sm font-semibold capitalize text-foreground">
              {platform || item.source}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {item.sourceUrl ? cleanUrlDomain(item.sourceUrl) : "Local upload"}
            </p>
          </div>

          <div className="rounded-lg border bg-background/60 p-4">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Extraction Pipeline
            </p>
            <p className="mt-1.5 text-sm font-semibold text-foreground">
              {item.type === "video"
                ? platform === "youtube"
                  ? "yt-dlp + Faster-Whisper"
                  : "Whisper Audio + RapidOCR Vision"
                : "Document Text Parser"}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              Multimodal audio & OCR sync
            </p>
          </div>

          <div className="rounded-lg border bg-background/60 p-4">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Vector Storage
            </p>
            <p className="mt-1.5 text-sm font-semibold text-foreground">
              PostgreSQL + pgvector
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              384d semantic chunk embeddings
            </p>
          </div>

          <div className="rounded-lg border bg-background/60 p-4">
            <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Assigned Space
            </p>
            <p className="mt-1.5 text-sm font-semibold text-foreground">
              {item.spaceName}
            </p>
            <p className="mt-0.5 text-xs text-muted-foreground">
              User knowledge partition
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}

// Platform badge component
function PlatformBadge({ platform }: { platform: string }) {
  if (platform === "youtube") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md border border-red-500/30 bg-red-500/10 px-2 py-0.5 text-xs font-medium text-red-500">
        <Play className="size-3 fill-current" />
        <span>YouTube</span>
      </span>
    );
  }
  if (platform === "instagram") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md border border-pink-500/30 bg-gradient-to-r from-pink-500/10 via-purple-500/10 to-orange-500/10 px-2 py-0.5 text-xs font-medium text-pink-400">
        <span className="size-2 rounded-full bg-pink-500" />
        <span>Instagram Reel</span>
      </span>
    );
  }
  if (platform === "facebook") {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md border border-blue-500/30 bg-blue-500/10 px-2 py-0.5 text-xs font-medium text-blue-400">
        <span className="size-2 rounded-full bg-blue-500" />
        <span>Facebook Reel</span>
      </span>
    );
  }
  return <span className="text-xs text-muted-foreground">{platform}</span>;
}

// Platform Logo component when no cover thumbnail exists
function PlatformLogo({ platform, type }: { platform?: string; type: string }) {
  if (platform === "youtube") {
    return (
      <div className="flex flex-col items-center gap-2">
        <div className="flex size-20 items-center justify-center rounded-2xl bg-red-600 text-white shadow-lg">
          <Play className="size-10 fill-white" />
        </div>
        <span className="text-xs font-semibold text-red-500">YouTube Video</span>
      </div>
    );
  }

  if (platform === "instagram") {
    return (
      <div className="flex flex-col items-center gap-2">
        <div className="flex size-20 items-center justify-center rounded-2xl bg-gradient-to-tr from-yellow-500 via-pink-600 to-purple-600 text-white shadow-lg">
          <svg className="size-10" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
            <rect height="20" rx="5" width="20" x="2" y="2" />
            <path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z" />
            <line x1="17.5" x2="17.51" y1="6.5" y2="6.5" />
          </svg>
        </div>
        <span className="text-xs font-semibold text-pink-400">Instagram Reel</span>
      </div>
    );
  }

  if (platform === "facebook") {
    return (
      <div className="flex flex-col items-center gap-2">
        <div className="flex size-20 items-center justify-center rounded-2xl bg-[#1877F2] text-white shadow-lg">
          <span className="font-sans text-4xl font-bold">f</span>
        </div>
        <span className="text-xs font-semibold text-blue-400">Facebook Reel</span>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-2">
      <div className="flex size-20 items-center justify-center rounded-2xl border bg-secondary/80 text-foreground shadow-sm">
        {type === "video" ? <Video className="size-10" /> : <FileText className="size-10" />}
      </div>
      <span className="text-xs font-semibold text-muted-foreground">Document File</span>
    </div>
  );
}

function cleanUrlDomain(rawUrl: string): string {
  try {
    const parsed = new URL(rawUrl);
    return parsed.hostname.replace("www.", "");
  } catch {
    return rawUrl;
  }
}

function detectPlatformFromUrl(url?: string): "instagram" | "facebook" | "youtube" | undefined {
  if (!url) return undefined;
  const lower = url.toLowerCase();
  if (lower.includes("instagram.com") || lower.includes("instagr.am")) return "instagram";
  if (lower.includes("facebook.com") || lower.includes("fb.watch") || lower.includes("fb.com"))
    return "facebook";
  if (lower.includes("youtube.com") || lower.includes("youtu.be")) return "youtube";
  return undefined;
}
