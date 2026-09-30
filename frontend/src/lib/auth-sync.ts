import { syncUser } from "@/lib/api";

type ClerkEmail = {
  emailAddress?: string | null;
};

type ClerkLikeUser = {
  id: string;
  fullName?: string | null;
  firstName?: string | null;
  lastName?: string | null;
  username?: string | null;
  primaryEmailAddress?: ClerkEmail | null;
  emailAddresses?: ClerkEmail[];
};

export function backendUserPayload(user: ClerkLikeUser) {
  const name =
    user.fullName ??
    (`${user.firstName ?? ""} ${user.lastName ?? ""}`.trim() || user.username) ??
    "Unknown";
  const email =
    user.primaryEmailAddress?.emailAddress ??
    user.emailAddresses?.[0]?.emailAddress ??
    "";

  return { id: user.id, name, email };
}

export async function syncBackendUser(user: ClerkLikeUser) {
  return syncUser(backendUserPayload(user));
}
