const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api";

export type IngestURLRequest = {
  user_id: string;
  space_id: string;
  url: string;
};

export type IngestResult = {
	content_id: string;
	ingestion_id: string;
	status: "pending" | "processing" | "completed" | "failed";
	title: string;
	content_type: "video" | "document" | "article" | "image";
	chunk_count: number;
};

export type IngestURLResult = IngestResult;

type APIResponse<T> = {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
};

export async function ingestURL(
	payload: IngestURLRequest
): Promise<IngestResult> {
  const response = await fetch(`${API_BASE_URL}/ingestion/url`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

	const body = (await response.json()) as APIResponse<IngestResult>;
  if (!response.ok || !body.success || !body.data) {
    throw new Error(body.error || body.message || "Ingestion failed");
  }

	return body.data;
}

export async function ingestFile(payload: {
	user_id: string;
	space_id: string;
	file: File;
}): Promise<IngestResult> {
	const formData = new FormData();
	formData.append("user_id", payload.user_id);
	formData.append("space_id", payload.space_id);
	formData.append("file", payload.file);

	const response = await fetch(`${API_BASE_URL}/ingestion/file`, {
		method: "POST",
		body: formData,
	});

	const body = (await response.json()) as APIResponse<IngestResult>;
	if (!response.ok || !body.success || !body.data) {
		throw new Error(body.error || body.message || "Ingestion failed");
	}

	return body.data;
}

// Why this file exists:
// The frontend should call one typed API helper instead of scattering fetch details
// through components. This mirrors the backend service boundary in TypeScript.
