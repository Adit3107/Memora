"use client";

import { FormEvent, useEffect, useState } from "react";
import { BookOpen, Loader2, Plus } from "lucide-react";

import { EmptyState } from "@/components/content/empty-state";
import { ErrorState } from "@/components/feedback/error-state";
import { LoadingState } from "@/components/feedback/loading-state";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import {
  createSpace,
  listSpaces,
  MEMORA_DEMO_USER_ID,
  type BackendSpace,
} from "@/lib/api";

export default function SpacesPage() {
  const [spaces, setSpaces] = useState<BackendSpace[]>([]);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");

  async function loadSpaces() {
    setIsLoading(true);
    setError("");
    try {
      setSpaces(await listSpaces());
    } catch {
      setError("Could not load spaces. Check that the Go backend is running.");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadSpaces();
  }, []);

  async function handleCreateSpace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmedName = name.trim();
    if (!trimmedName) {
      setError("Add a Space name before creating it.");
      return;
    }

    setIsSaving(true);
    setError("");
    try {
      const created = await createSpace({
        user_id: MEMORA_DEMO_USER_ID,
        name: trimmedName,
        description: description.trim(),
      });
      setSpaces((current) => [created, ...current]);
      setName("");
      setDescription("");
    } catch {
      setError("Space could not be created. Check the backend and try again.");
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader
          eyebrow="Spaces"
          title="Topic rooms for everything you save."
          description="Create focused collections, then choose the right Space while saving videos and documents."
        />
      </div>

      <form
        className="rounded-md border bg-card p-5 shadow-sm"
        onSubmit={handleCreateSpace}
      >
        <h2 className="text-lg font-semibold">Create Space</h2>
        <div className="mt-5 grid gap-3 md:grid-cols-[1fr_1.4fr_auto]">
          <input
            className="h-10 rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
            disabled={isSaving}
            onChange={(event) => setName(event.target.value)}
            placeholder="Space name"
            value={name}
          />
          <input
            className="h-10 rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
            disabled={isSaving}
            onChange={(event) => setDescription(event.target.value)}
            placeholder="Description"
            value={description}
          />
          <Button disabled={isSaving || !name.trim()} type="submit">
            {isSaving ? <Loader2 className="size-4 animate-spin" /> : <Plus className="size-4" />}
            Create
          </Button>
        </div>
      </form>

      {error ? (
        <ErrorState
          description={error}
          onRetry={() => void loadSpaces()}
          title="Spaces unavailable"
        />
      ) : null}

      {isLoading ? (
        <LoadingState title="Loading spaces" description="Fetching your Spaces from the backend." />
      ) : null}

      {!isLoading && !error && spaces.length > 0 ? (
        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {spaces.map((space) => (
            <article
              className="rounded-md border bg-card p-5 shadow-sm transition-colors hover:bg-accent/60"
              key={space.id}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <h2 className="truncate text-lg font-semibold">{space.name}</h2>
                  <p className="mt-2 line-clamp-3 text-sm leading-6 text-muted-foreground">
                    {space.description || "No description yet."}
                  </p>
                </div>
                <span className="flex size-10 shrink-0 items-center justify-center rounded-md border bg-background">
                  <BookOpen className="size-5 text-muted-foreground" aria-hidden="true" />
                </span>
              </div>
              <p className="mt-4 text-xs text-muted-foreground">
                Created {formatDate(space.created_at)}
              </p>
            </article>
          ))}
        </section>
      ) : null}

      {!isLoading && !error && spaces.length === 0 ? (
        <EmptyState
          title="No spaces yet"
          description="Create your first Space to group saved videos, documents, articles, and images."
        />
      ) : null}
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(new Date(value));
}
