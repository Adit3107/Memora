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

export type SemanticSearchRequest = {
	user_id: string;
	space_id?: string;
	query: string;
	limit?: number;
};

export type SearchMode = "semantic" | "keyword" | "hybrid";

export type SearchRequest = {
	user_id: string;
	query: string;
	mode?: SearchMode;
	space_id?: string;
	content_type?: "video" | "document" | "article" | "image";
	source_type?: string;
	tag_ids?: string[];
	created_from?: string;
	created_to?: string;
	limit?: number;
	offset?: number;
};

export type SearchResult = {
	content_id: string;
	chunk_id: string;
	chunk_index: number;
	title: string;
	content_type: "video" | "document" | "article" | "image";
	source_url?: string;
	thumbnail_url?: string;
	text: string;
	score: number;
	page_index?: number;
	start_seconds?: number;
	end_seconds?: number;
	source_type: string;
	embedding_model: string;
	metadata: Record<string, string>;
	tags: string[];
};

export type SearchResponse = {
	mode: SearchMode;
	query: string;
	total: number;
	results: SearchResult[];
};

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

export async function semanticSearch(
	payload: SemanticSearchRequest
): Promise<SearchResult[]> {
	const response = await fetch(`${API_BASE_URL}/search/semantic`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(payload),
	});

	const body = (await response.json()) as APIResponse<SearchResult[]>;
	if (!response.ok || !body.success || !body.data) {
		throw new Error(body.error || body.message || "Semantic search failed");
	}

	return body.data;
}

export async function searchMemora(
	payload: SearchRequest
): Promise<SearchResponse> {
	const response = await fetch(`${API_BASE_URL}/search`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
		},
		body: JSON.stringify(payload),
	});

	const body = (await response.json()) as APIResponse<SearchResponse>;
	if (!response.ok || !body.success || !body.data) {
		throw new Error(body.error || body.message || "Search failed");
	}

	return body.data;
}

// Why this file exists:
// The frontend should call one typed API helper instead of scattering fetch details
// through components. This mirrors the backend service boundary in TypeScript.
