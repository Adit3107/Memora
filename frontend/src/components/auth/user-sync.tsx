"use client";

import { useUser } from "@clerk/nextjs";
import { useEffect, useRef } from "react";
import { syncUser } from "@/lib/api";

/**
 * Invisible component mounted inside the /app layout.
 * On first render (or when the Clerk user changes) it calls POST /users/sync
 * so the Go backend always has an up-to-date record for the logged-in user.
 */
export function UserSync() {
  const { user, isLoaded } = useUser();
  const lastSyncedId = useRef<string | null>(null);

  useEffect(() => {
    if (!isLoaded || !user) return;
    // Only sync once per session per user ID
    if (lastSyncedId.current === user.id) return;
    lastSyncedId.current = user.id;

    const name =
      user.fullName ??
      (`${user.firstName ?? ""} ${user.lastName ?? ""}`.trim() || user.username) ??
      "Unknown";
    const email =
      user.primaryEmailAddress?.emailAddress ??
      user.emailAddresses[0]?.emailAddress ??
      "";

    void syncUser({ id: user.id, name, email })
      .then(() => window.dispatchEvent(new Event("mindshelf:spaces-changed")))
      .catch(() => {
        // Non-fatal: if sync fails the user can still browse; will retry on next load
      });
  }, [isLoaded, user]);

  return null;
}
