"use client";

import { CheckCircle2 } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";

import {
  saveInputOptions,
  suggestedTags,
  supportedFileFormats,
} from "@/data/save-content";
import { spaces } from "@/data/content";
import { cn } from "@/lib/utils";
import type { SaveInputType } from "@/types/save-content";

import { Button } from "../ui/button";

const fileTypes: SaveInputType[] = ["file", "image"];

export function SaveContentForm() {
  const [inputType, setInputType] = useState<SaveInputType>("youtube");
  const [url, setUrl] = useState("");
  const [spaceSlug, setSpaceSlug] = useState(spaces[0]?.slug ?? "");
  const [selectedTags, setSelectedTags] = useState<string[]>(["RAG", "AI"]);
  const [fileName, setFileName] = useState("");
  const [message, setMessage] = useState("");

  const isFileInput = fileTypes.includes(inputType);
  const selectedSpace = useMemo(
    () => spaces.find((space) => space.slug === spaceSlug),
    [spaceSlug]
  );

  function toggleTag(tag: string) {
    setSelectedTags((current) =>
      current.includes(tag)
        ? current.filter((item) => item !== tag)
        : [...current, tag]
    );
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (isFileInput && !fileName) {
      setMessage("Choose a file placeholder before continuing.");
      return;
    }

    if (!isFileInput && !url.trim()) {
      setMessage("Enter a URL before continuing.");
      return;
    }

    setMessage(
      "Phase 1 preview only. Nothing was uploaded, saved, processed, embedded, or sent to a backend."
    );
  }

  return (
    <form className="space-y-6" onSubmit={handleSubmit}>
      <section className="grid gap-3 md:grid-cols-5" aria-label="Input type">
        {saveInputOptions.map((option) => {
          const Icon = option.icon;
          const isActive = inputType === option.value;

          return (
            <button
              className={cn(
                "rounded-md border bg-card p-4 text-left shadow-sm transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50",
                isActive && "border-primary bg-accent text-accent-foreground"
              )}
              key={option.value}
              onClick={() => {
                setInputType(option.value);
                setMessage("");
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
            {isFileInput ? (
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="file-placeholder">
                  Upload {inputType === "image" ? "image" : "file"}
                </label>
                <input
                  className="block w-full rounded-md border bg-background px-3 py-2 text-sm file:mr-3 file:rounded-md file:border-0 file:bg-secondary file:px-3 file:py-1 file:text-sm file:font-medium"
                  id="file-placeholder"
                  onChange={(event) =>
                    setFileName(event.target.files?.[0]?.name ?? "")
                  }
                  type="file"
                />
                <p className="text-xs leading-5 text-muted-foreground">
                  This does not upload anything. It only previews the future
                  file-selection flow.
                </p>
              </div>
            ) : (
              <div className="space-y-2">
                <label className="text-sm font-medium" htmlFor="content-url">
                  URL
                </label>
                <input
                  className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                  id="content-url"
                  onChange={(event) => setUrl(event.target.value)}
                  placeholder="https://..."
                  type="url"
                  value={url}
                />
              </div>
            )}

            <div className="grid gap-5 md:grid-cols-2">
              <label className="space-y-2 text-sm font-medium">
                Space
                <select
                  className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                  onChange={(event) => setSpaceSlug(event.target.value)}
                  value={spaceSlug}
                >
                  {spaces.map((space) => (
                    <option key={space.slug} value={space.slug}>
                      {space.name}
                    </option>
                  ))}
                </select>
              </label>

              <div className="space-y-2">
                <p className="text-sm font-medium">Selected destination</p>
                <div className="rounded-md border bg-background px-3 py-2 text-sm text-muted-foreground">
                  {selectedSpace?.name ?? "No Space selected"}
                </div>
              </div>
            </div>

            <div className="space-y-3">
              <p className="text-sm font-medium">Tags</p>
              <div className="flex flex-wrap gap-2">
                {suggestedTags.map((tag) => {
                  const isActive = selectedTags.includes(tag);

                  return (
                    <button
                      className={cn(
                        "rounded-md border px-3 py-2 text-sm font-medium transition-colors",
                        isActive
                          ? "bg-primary text-primary-foreground"
                          : "bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                      )}
                      key={tag}
                      onClick={() => toggleTag(tag)}
                      type="button"
                    >
                      #{tag}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>

          <aside className="rounded-md border bg-background p-4">
            <h2 className="text-sm font-semibold">Supported formats</h2>
            <div className="mt-3 flex flex-wrap gap-2">
              {supportedFileFormats.map((format) => (
                <span
                  className="rounded-md bg-secondary px-2 py-1 text-xs font-medium text-secondary-foreground"
                  key={format}
                >
                  {format}
                </span>
              ))}
            </div>
            <p className="mt-4 text-sm leading-6 text-muted-foreground">
              Future phases will extract transcripts, text, OCR, chunks,
              embeddings, metadata, and citations. Phase 1 only represents the
              workflow.
            </p>
          </aside>
        </div>

        <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center">
          <Button type="submit">Save to Memora</Button>
          <p className="text-sm text-muted-foreground">
            Mock action. No backend request is made.
          </p>
        </div>

        {message ? (
          <div className="mt-5 flex gap-3 rounded-md border bg-background p-4 text-sm">
            <CheckCircle2 className="size-5 shrink-0 text-muted-foreground" />
            <p className="leading-6 text-muted-foreground">{message}</p>
          </div>
        ) : null}
      </section>
    </form>
  );
}
