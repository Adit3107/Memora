"use client";

import { SignOutButton, useUser } from "@clerk/nextjs";
import {
  Calendar,
  CheckCircle2,
  Copy,
  Key,
  Layers,
  Lock,
  Mail,
  RefreshCw,
  Save,
  ShieldCheck,
  Sparkles,
  User,
  X,
} from "lucide-react";
import Link from "next/link";
import type { FormEvent } from "react";
import { useEffect, useState } from "react";

import { listContent, listSpaces } from "@/lib/api";

export default function ProfilePage() {
  const { isLoaded, user } = useUser();
  const [copied, setCopied] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [profileError, setProfileError] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [contentCount, setContentCount] = useState<number | null>(null);
  const [spaceCount, setSpaceCount] = useState<number | null>(null);

  useEffect(() => {
    async function fetchStats() {
      if (!user?.id) return;
      try {
        const [c, s] = await Promise.all([
          listContent(user.id),
          listSpaces(user.id),
        ]);
        setContentCount(c.length);
        setSpaceCount(s.length);
      } catch {
        setContentCount(0);
        setSpaceCount(0);
      }
    }
    void fetchStats();
  }, [user?.id]);

  async function saveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!user) return;

    setIsSavingProfile(true);
    setProfileError("");
    try {
      await user.update({
        firstName: firstName.trim(),
        lastName: lastName.trim(),
      });
      await user.reload();
      setIsEditing(false);
    } catch {
      setProfileError("Profile could not be updated. Please try again.");
    } finally {
      setIsSavingProfile(false);
    }
  }

  function copyUserId() {
    if (!user?.id) return;
    void navigator.clipboard.writeText(user.id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  if (!isLoaded) {
    return (
      <div className="flex min-h-[400px] items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <RefreshCw className="size-6 animate-spin text-primary" />
          <p className="text-sm text-muted-foreground">Loading Mindshelf profile...</p>
        </div>
      </div>
    );
  }

  const primaryEmail =
    user?.primaryEmailAddress?.emailAddress ||
    user?.emailAddresses?.[0]?.emailAddress ||
    "user@mindshelf.app";

  const isEmailVerified =
    user?.primaryEmailAddress?.verification?.status === "verified";

  const createdAtFormatted = user?.createdAt
    ? new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(new Date(user.createdAt))
    : "Recently";

  const lastSignInFormatted = user?.lastSignInAt
    ? new Intl.DateTimeFormat(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(new Date(user.lastSignInAt))
    : "Current session";

  const externalGoogle = user?.externalAccounts?.find(
    (acc) => acc.provider === "google"
  );

  return (
    <div className="space-y-8">
      {/* Top Banner / Heading */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Account & Security
          </p>
          <h1 className="mt-1 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            User Profile
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Manage your authenticated identity, verification status, and workspace details.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <button
            className="inline-flex items-center gap-2 rounded-lg border border-primary/30 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition hover:bg-primary/20"
            onClick={() => {
              if (!isEditing) {
                setFirstName(user?.firstName ?? "");
                setLastName(user?.lastName ?? "");
              }
              setProfileError("");
              setIsEditing((current) => !current);
            }}
            type="button"
          >
            {isEditing ? <X className="size-3.5" /> : <User className="size-3.5" />}
            {isEditing ? "Cancel" : "Edit Profile"}
          </button>
          <SignOutButton>
            <button className="rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-2 text-sm font-medium text-destructive transition hover:bg-destructive/20">
              Sign out
            </button>
          </SignOutButton>
          <Link
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm transition hover:bg-primary/90"
            href="/app"
          >
            Back to Home
          </Link>
        </div>
      </div>

      {isEditing ? (
        <section className="rounded-xl border border-primary/20 bg-card p-6 shadow-sm">
          <form className="grid gap-4 sm:grid-cols-[1fr_1fr_auto] sm:items-end" onSubmit={saveProfile}>
            <label className="space-y-2 text-sm font-medium">
              First name
              <input
                className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                onChange={(event) => setFirstName(event.target.value)}
                value={firstName}
              />
            </label>
            <label className="space-y-2 text-sm font-medium">
              Last name
              <input
                className="h-10 w-full rounded-md border bg-background px-3 text-sm outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                onChange={(event) => setLastName(event.target.value)}
                value={lastName}
              />
            </label>
            <button
              className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground disabled:opacity-60"
              disabled={isSavingProfile}
              type="submit"
            >
              <Save className="size-4" />
              {isSavingProfile ? "Saving" : "Save changes"}
            </button>
          </form>
          {profileError ? <p className="mt-3 text-sm text-destructive">{profileError}</p> : null}
        </section>
      ) : null}

      {/* Main Profile Identity Card */}
      <section className="rounded-2xl border border-border/80 bg-card p-6 shadow-sm sm:p-8">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-5">
            {user?.imageUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                alt={user.fullName || "User Avatar"}
                className="size-20 rounded-full border-2 border-primary/30 object-cover shadow-md"
                src={user.imageUrl}
              />
            ) : (
              <div className="flex size-20 items-center justify-center rounded-full bg-gradient-to-tr from-primary to-primary/60 text-2xl font-bold text-primary-foreground shadow-md">
                {user?.firstName?.[0] || user?.fullName?.[0] || "M"}
              </div>
            )}

            <div className="space-y-1">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="text-2xl font-bold tracking-tight text-foreground">
                  {user?.fullName || user?.firstName || "Mindshelf Member"}
                </h2>
                <span className="rounded-full border border-primary/30 bg-primary/10 px-2.5 py-0.5 text-xs font-semibold text-primary">
                  Pro Plan
                </span>
              </div>
              <p className="text-sm font-medium text-muted-foreground">{primaryEmail}</p>
              <div className="flex items-center gap-2 pt-1">
                <span className="inline-flex items-center gap-1 rounded-md border border-emerald-500/30 bg-emerald-500/10 px-2 py-0.5 text-xs font-medium text-emerald-400">
                  <CheckCircle2 className="size-3" />
                  <span>Authenticated</span>
                </span>
                {isEmailVerified ? (
                  <span className="inline-flex items-center gap-1 rounded-md border border-blue-500/30 bg-blue-500/10 px-2 py-0.5 text-xs font-medium text-blue-400">
                    <ShieldCheck className="size-3" />
                    <span>Email Verified</span>
                  </span>
                ) : null}
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-2 rounded-xl border border-border/60 bg-background/60 p-4 sm:w-72">
            <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Clerk User Identifier
            </span>
            <div className="flex items-center justify-between gap-2">
              <span className="font-mono text-xs text-foreground truncate">
                {user?.id || "user_demo"}
              </span>
              <button
                className="inline-flex items-center gap-1 rounded border border-border px-2 py-1 text-xs font-medium text-muted-foreground hover:text-foreground"
                onClick={copyUserId}
                title="Copy User ID"
                type="button"
              >
                <Copy className="size-3" />
                <span>{copied ? "Copied" : "Copy"}</span>
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* 3 Detail Grids: Account Details, Security & Verification, Workspace Stats */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Card 1: Account Information */}
        <section className="rounded-xl border border-border/80 bg-card p-6 shadow-sm space-y-4">
          <div className="flex items-center gap-2.5 border-b pb-3">
            <User className="size-4 text-primary" />
            <h3 className="text-base font-semibold text-foreground">Account Information</h3>
          </div>

          <div className="space-y-3 text-sm">
            <div>
              <p className="text-xs text-muted-foreground">Full Name</p>
              <p className="font-medium text-foreground">
                {user?.fullName || `${user?.firstName || ""} ${user?.lastName || ""}`.trim() || "Not specified"}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Username</p>
              <p className="font-medium text-foreground">
                {user?.username ? `@${user.username}` : "None"}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Member Since</p>
              <p className="font-medium text-foreground">{createdAtFormatted}</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Last Login Activity</p>
              <p className="font-medium text-foreground">{lastSignInFormatted}</p>
            </div>
          </div>
        </section>

        {/* Card 2: Security & Email Verification */}
        <section className="rounded-xl border border-border/80 bg-card p-6 shadow-sm space-y-4">
          <div className="flex items-center gap-2.5 border-b pb-3">
            <Lock className="size-4 text-primary" />
            <h3 className="text-base font-semibold text-foreground">Security & Verification</h3>
          </div>

          <div className="space-y-3 text-sm">
            <div>
              <p className="text-xs text-muted-foreground">Primary Email Address</p>
              <div className="flex items-center justify-between">
                <span className="font-medium text-foreground">{primaryEmail}</span>
                {isEmailVerified ? (
                  <span className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-400">
                    VERIFIED
                  </span>
                ) : (
                  <span className="rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-amber-400">
                    PENDING
                  </span>
                )}
              </div>
            </div>

            <div>
              <p className="text-xs text-muted-foreground">Authentication Method</p>
              <p className="font-medium text-foreground">
                {externalGoogle
                  ? "Google OAuth 2.0"
                  : user?.passwordEnabled
                  ? "Email & Password + OTP"
                  : "Email Verification OTP"}
              </p>
            </div>

            <div>
              <p className="text-xs text-muted-foreground">Two-Factor Authentication (2FA)</p>
              <p className="font-medium text-foreground">
                {user?.twoFactorEnabled ? "Enabled" : "Disabled (Optional)"}
              </p>
            </div>

            <div>
              <p className="text-xs text-muted-foreground">Active Session</p>
              <span className="inline-flex items-center gap-1.5 font-medium text-emerald-400">
                <span className="size-2 rounded-full bg-emerald-500 animate-pulse" />
                Live Session Connected
              </span>
            </div>
          </div>
        </section>

        {/* Card 3: Workspace & Vector Storage */}
        <section className="rounded-xl border border-border/80 bg-card p-6 shadow-sm space-y-4">
          <div className="flex items-center gap-2.5 border-b pb-3">
            <Sparkles className="size-4 text-primary" />
            <h3 className="text-base font-semibold text-foreground">Workspace Metrics</h3>
          </div>

          <div className="space-y-3 text-sm">
            <div>
              <p className="text-xs text-muted-foreground">Saved Videos & Documents</p>
              <p className="text-lg font-bold text-foreground">
                {contentCount !== null ? contentCount : "..."} items
              </p>
            </div>

            <div>
              <p className="text-xs text-muted-foreground">Active Knowledge Spaces</p>
              <p className="text-lg font-bold text-foreground">
                {spaceCount !== null ? spaceCount : "..."} spaces
              </p>
            </div>

            <div>
              <p className="text-xs text-muted-foreground">Vector Search Engine</p>
              <p className="font-medium text-foreground">PostgreSQL + pgvector (384d)</p>
            </div>

            <div>
              <p className="text-xs text-muted-foreground">AI Summarization</p>
              <p className="font-medium text-foreground">Gemini 3.8 Flash</p>
            </div>
          </div>
        </section>
      </div>

      {/* Account actions section */}
      <section className="rounded-xl border border-border/80 bg-card p-6 shadow-sm">
        <h3 className="text-base font-semibold text-foreground">Session Control</h3>
        <p className="mt-1 text-xs text-muted-foreground">
          Sign out of your active Mindshelf session across all devices.
        </p>

        <div className="mt-4 flex flex-wrap gap-3">
          <SignOutButton>
            <button className="rounded-lg bg-destructive px-5 py-2.5 text-sm font-semibold text-destructive-foreground shadow-sm transition hover:bg-destructive/90">
              Sign out of Mindshelf
            </button>
          </SignOutButton>

          <Link
            className="rounded-lg border border-border bg-background px-5 py-2.5 text-sm font-medium text-foreground transition hover:bg-accent"
            href="/app/settings"
          >
            Workspace Settings
          </Link>
        </div>
      </section>
    </div>
  );
}
