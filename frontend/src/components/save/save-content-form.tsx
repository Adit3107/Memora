"use client";

import {
  AlertCircle,
  CheckCircle2,
  FileText,
  Loader2,
  Plus,
  UploadCloud,
  Video,
} from "lucide-react";
import Link from "next/link";
import { FormEvent, useEffect, useMemo, useState } from "react";

import {
  createTag,
  getIngestionResult,
  ingestFileWithProgress,
  ingestURL,
  listSpaces,
  listTags,
  MEMORA_DEMO_SPACE_ID,
  MEMORA_DEMO_USER_ID,
  setContentTags,
  type BackendSpace,
  type BackendTag,
  type IngestResult,
} from "@/lib/api";
import { cn } from "@/lib/utils";

import { Button } from "../ui/button";

type InputType = "video" | "document";
type SubmitState = "idle" | "processing" | "success" | "error";

const inputOptions: {
  value: InputType;
  label: string;
  description: string;
  icon: typeof Video;
}[] = [
  {
    value: "video",
    label: "Video",
    description: "YouTube URL",
    icon: Video,
  },
  {
    value: "document",
    label: "Document",
    description: "PDF, DOCX, PPTX, TXT",
    icon: FileText,
  },
];

const processingStages = [
  "Content received",
  "Extracting content",
  "Generating embeddings",
  "Finalizing library entry",
];

export function SaveContentForm() {
  const [inputType, setInputType] = useState<InputType>("video");
  const [contentName, setContentName] = useState("");
  const [url, setUrl] = useState("");
  const [spaces, setSpaces] = useState<BackendSpace[]>([]);
  const [availableTags, setAvailableTags] = useState<BackendTag[]>([]);
  const [spaceID, setSpaceID] = useState(MEMORA_DEMO_SPACE_ID);
  const [selectedTagNames, setSelectedTagNames] = useState<string[]>([]);
  const [newTag, setNewTag] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [state, setState] = useState<SubmitState>("idle");
  const [error, setError] = useState("");
  const [result, setResult] = useState<IngestResult | null>(null);
  const [stageIndex, setStageIndex] = useState(0);
  const [uploadProgress, setUploadProgress] = useState(0);

  const isProcessing = state === "processing";
  const selectedSpace = useMemo(
    () => spaces.find((space) => space.id === spaceID),
    [spaceID, spaces]
  );

  async function loadFormData() {
    try {
      const [spaceRows, tagRows] = await Promise.all([listSpaces(), listTags()]);
      setSpaces(spaceRows);
      setAvailableTags(tagRows);
      setSpaceID(
        spaceRows.find((space) => space.id === MEMORA_DEMO_SPACE_ID)?.id ??
          spaceRows[0]?.id ??
          MEMORA_DEMO_SPACE_ID
      );
      setSelectedTagNames(tagRows.slice(0, 2).map((tag) => tag.name));
    } catch {
      setError("Could not load spaces and tags. Check that the Go backend is running.");
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadFormData();
  }, []);

  useEffect(() => {
    if (!isProcessing) {
      return;
    }

    const interval = window.setInterval(() => {
      setStageIndex((current) =>
        current >= processingStages.length - 1 ? current : current + 1
      );
    }, 1300);

    return () => window.clearInterval(interval);
  }, [isProcessing]);

  function toggleTag(tagName: string) {
    setSelectedTagNames((current) =>
      current.includes(tagName)
        ? current.filter((item) => item !== tagName)
        : [...current, tagName]
    );
  }

  function addTypedTag() {
    const trimmed = newTag.trim();
    if (!trimmed) {
      return;
    }
    setSelectedTagNames((current) =>
      current.some((tag) => tag.toLowerCase() === trimmed.toLowerCase())
        ? current
        : [...current, trimmed]
    );
    setNewTag("");
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setResult(null);
    setUploadProgress(0);

    if (!spaceID.trim()) {
      setError("Select a Space before saving.");
      setState("error");
      return;
    }

    if (inputType === "video" && !url.trim()) {
      setError("Paste a YouTube URL before saving.");
      setState("error");
      return;
    }

    if (inputType === "document" && !selectedFile) {
      setError("Choose a document before uploading.");
      setState("error");
      return;
    }

    setStageIndex(0);
    setState("processing");
    try {
      const saved =
        inputType === "document" && selectedFile
          ? await ingestFileWithProgress(
              {
                user_id: MEMORA_DEMO_USER_ID,
                space_id: spaceID,
                name: contentName.trim() || undefined,
                file: selectedFile,
              },
              setUploadProgress
            )
          : await ingestURL({
              user_id: MEMORA_DEMO_USER_ID,
              space_id: spaceID,
              name: contentName.trim() || undefined,
              url: url.trim(),
            });

      const finalResult =
        saved.status === "pending" || saved.status === "processing"
          ? await pollIngestion(saved.ingestion_id)
          : saved;

      if (selectedTagNames.length > 0) {
        await applyTags(finalResult.content_id, selectedTagNames);
      }

      setResult(finalResult);
      setStageIndex(processingStages.length - 1);
      setState("success");
    } catch (caught) {
      console.error(caught);
      setError("We couldn't process this content. Please try again.");
      setState("error");
    }
  }

  async function applyTags(contentID: string, tagNames: string[]) {
    const existingTags = await listTags();
    const ensuredTags = await Promise.all(
      tagNames.map(async (tagName) => {
        const existing = existingTags.find(
          (tag) => tag.name.toLowerCase() === tagName.toLowerCase()
        );
        return (
          existing ??
          createTag({
            user_id: MEMORA_DEMO_USER_ID,
            name: tagName,
          })
        );
      })
    );
    await setContentTags(
      contentID,
      ensuredTags.map((tag) => tag.id)
    );
    setAvailableTags(await listTags());
  }

  async function pollIngestion(ingestionID: string) {
    let latest = await getIngestionResult(ingestionID);
    for (let attempt = 0; attempt < 8; attempt += 1) {
      if (latest.status === "completed" || latest.status === "failed") {
        return latest;
      }
      await new Promise((resolve) => window.setTimeout(resolve, 1200));
      latest = await getIngestionResult(ingestionID);
    }
    return latest;
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit}>
      <section className="grid gap-3 sm:grid-cols-2" aria-label="Input type">
        {inputOptions.map((option) => {
          const Icon = option.icon;
          const isActive = inputType === option.value;

          return (
            <button
              className={cn(
                "rounded-md border bg-card p-4 text-left shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50",
                isActive && "border-primary bg-accent text-accent-foreground"
              )}
              disabled={isProcessing}
              key={option.value}
              onClick={() => {
                setInputType(option.value);
                setState("idle");
                setError("");
                setResult(null);
              }}
              type="button"
            >
              <Icon className="size-5 text-muted-foreground" aria-hidden="true" />
              <span className="mt-3 block text-sm font-semibold">
                {option.label}
              </span>
              <span className="mt-1 block text-xs leading-5 text-muted-foreground">
                {option.description}
              </span>
            </button>
          );
        })}
      </section>

      <section className="rounded-md border bg-card p-5 shadow-sm">
        <div className="grid gap-5 lg:grid-cols-[1.2fr_0.8fr]">
          <div className="space-y-5">
            <label className="space-y-2 text-sm font-medium" htmlFor="content-name">
              Content name
              <input
                className="h-11 w-full rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                disabled={isProcessing}
                id="content-name"
                onChange={(event) => setContentName(event.target.value)}
                placeholder={
                  inputType === "video"
                    ? "Kafka cab-booking short"
                    : "Blockchain unit 2 notes"
                }
                type="text"
                value={contentName}
              />
            </label>

            {inputType === "video" ? (
              <label className="space-y-2 text-sm font-medium" htmlFor="content-url">
                Paste a YouTube URL
                <input
                  className="h-11 w-full rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                  disabled={isProcessing}
                  id="content-url"
                  onChange={(event) => setUrl(event.target.value)}
                  placeholder="https://youtube.com/watch?v=..."
                  type="url"
                  value={url}
                />
              </label>
            ) : (
              <label
                className="flex min-h-52 cursor-pointer flex-col items-center justify-center rounded-md border border-dashed bg-background p-6 text-center transition-colors hover:bg-accent/40"
                htmlFor="content-file"
              >
                <UploadCloud className="size-8 text-muted-foreground" aria-hidden="true" />
                <span className="mt-4 text-sm font-semibold">
                  {selectedFile ? selectedFile.name : "Drop your document here"}
                </span>
                <span className="mt-1 text-sm text-muted-foreground">
                  or click to browse
                </span>
                <span className="mt-4 text-xs font-medium text-muted-foreground">
                  PDF, DOCX, PPTX, TXT, CSV, XLSX
                </span>
                <input
                  accept=".pdf,.doc,.docx,.ppt,.pptx,.txt,.csv,.xlsx"
                  className="sr-only"
                  disabled={isProcessing}
                  id="content-file"
                  onChange={(event) => {
                    const file = event.target.files?.[0] ?? null;
                    setSelectedFile(file);
                    if (file && !contentName.trim()) {
                      setContentName(file.name);
                    }
                  }}
                  type="file"
                />
              </label>
            )}

            <label className="space-y-2 text-sm font-medium">
              Space
              <select
                className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                disabled={isProcessing}
                onChange={(event) => setSpaceID(event.target.value)}
                value={spaceID}
              >
                {spaces.length === 0 ? (
                  <option value={MEMORA_DEMO_SPACE_ID}>Demo Space</option>
                ) : (
                  spaces.map((space) => (
                    <option key={space.id} value={space.id}>
                      {space.name}
                    </option>
                  ))
                )}
              </select>
            </label>

            <div className="space-y-3">
              <p className="text-sm font-medium">Tags</p>
              <div className="flex flex-wrap gap-2">
                {availableTags.map((tag) => {
                  const isActive = selectedTagNames.includes(tag.name);

                  return (
                    <button
                      className={cn(
                        "rounded-md border px-3 py-2 text-sm font-medium transition-colors",
                        isActive
                          ? "bg-primary text-primary-foreground"
                          : "bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                      )}
                      disabled={isProcessing}
                      key={tag.id}
                      onClick={() => toggleTag(tag.name)}
                      type="button"
                    >
                      #{tag.name}
                    </button>
                  );
                })}
              </div>
              <div className="flex gap-2">
                <input
                  className="h-10 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                  disabled={isProcessing}
                  onChange={(event) => setNewTag(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      addTypedTag();
                    }
                  }}
                  placeholder="Add tag"
                  value={newTag}
                />
                <Button
                  disabled={isProcessing || !newTag.trim()}
                  onClick={addTypedTag}
                  type="button"
                  variant="outline"
                >
                  <Plus className="size-4" />
                  Add
                </Button>
              </div>
            </div>
          </div>

          <aside className="rounded-md border bg-background p-4">
            <h2 className="text-sm font-semibold">Destination</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              {selectedSpace
                ? `${selectedSpace.name}: ${selectedSpace.description || "No description"}`
                : "Content will be sent to the selected backend Space."}
            </p>

            {isProcessing ? (
              <div className="mt-5 space-y-4">
                <div className="flex items-center gap-2 text-sm font-medium">
                  <Loader2 className="size-4 animate-spin" />
                  {inputType === "document" && uploadProgress < 100
                    ? `Uploading... ${uploadProgress}%`
                    : "Processing your content"}
                </div>
                {inputType === "document" && uploadProgress < 100 ? (
                  <div className="space-y-2">
                    <div className="h-2 overflow-hidden rounded-full bg-muted">
                      <div
                        className="h-full rounded-full bg-primary transition-all"
                        style={{ width: `${uploadProgress}%` }}
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Upload progress is reported by the browser before Go begins
                      extraction and embedding work.
                    </p>
                  </div>
                ) : null}
                <div className="space-y-2">
                  {processingStages.map((stage, index) => (
                    <div className="flex items-center gap-2 text-sm" key={stage}>
                      {index < stageIndex ? (
                        <CheckCircle2 className="size-4 text-muted-foreground" />
                      ) : index === stageIndex ? (
                        <Loader2 className="size-4 animate-spin text-muted-foreground" />
                      ) : (
                        <span className="size-4 rounded-full border" />
                      )}
                      <span className="text-muted-foreground">{stage}</span>
                    </div>
                  ))}
                </div>
                <div className="h-2 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-primary transition-all"
                    style={{
                      width: `${((stageIndex + 1) / processingStages.length) * 100}%`,
                    }}
                  />
                </div>
              </div>
            ) : null}
          </aside>
        </div>

        <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center">
          <Button disabled={isProcessing} type="submit">
            {isProcessing ? (
              <>
                <Loader2 className="size-4 animate-spin" />
                Saving
              </>
            ) : (
              "Save Content"
            )}
          </Button>
          <p className="text-sm text-muted-foreground">
            Requests go to the Go backend. Go handles extraction, embeddings, and storage.
          </p>
        </div>

        {state === "success" && result ? (
          <div className="mt-5 flex gap-3 rounded-md border bg-background p-4 text-sm">
            <CheckCircle2 className="size-5 shrink-0 text-muted-foreground" />
            <div className="space-y-1 leading-6 text-muted-foreground">
              <p className="font-medium text-foreground">Saved to Memora</p>
              <p>
                &quot;{result.title}&quot; is {result.status} with {result.chunk_count} searchable
                chunk{result.chunk_count === 1 ? "" : "s"}.
              </p>
              <Link
                className="font-medium text-foreground hover:underline"
                href={`/app/library/${result.content_id}`}
              >
                Open content
              </Link>
            </div>
          </div>
        ) : null}

        {state === "error" && error ? (
          <div className="mt-5 flex gap-3 rounded-md border bg-background p-4 text-sm">
            <AlertCircle className="size-5 shrink-0 text-muted-foreground" />
            <div className="space-y-2">
              <p className="font-medium text-foreground">Something went wrong</p>
              <p className="leading-6 text-muted-foreground">{error}</p>
            </div>
          </div>
        ) : null}
      </section>
    </form>
  );
}
