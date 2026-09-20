const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api";

export type IngestURLRequest = {
  user_id: string;
  space_id: string;
  url: string;
};

export type IngestURLResult = {
  content_id: string;
  ingestion_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  title: string;
  content_type: "video" | "document" | "article" | "image";
  chunk_count: number;
};

type APIResponse<T> = {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
};

export async function ingestURL(
  payload: IngestURLRequest
): Promise<IngestURLResult> {
  const response = await fetch(`${API_BASE_URL}/ingestion/url`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  const body = (await response.json()) as APIResponse<IngestURLResult>;
  if (!response.ok || !body.success || !body.data) {
    throw new Error(body.error || body.message || "Ingestion failed");
  }

  return body.data;
}
