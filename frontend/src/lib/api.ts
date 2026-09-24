const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api";

export const MEMORA_DEMO_USER_ID =
	process.env.NEXT_PUBLIC_MEMORA_USER_ID ??
	"67a79aff-376d-48a7-af69-f087d46d313e";

export const MEMORA_DEMO_SPACE_ID =
	process.env.NEXT_PUBLIC_MEMORA_SPACE_ID ??
	"2e96706f-8134-448b-918d-979aeb0500bc";

export type ContentType = "video" | "document" | "article" | "image";

export type BackendContent = {
	id: string;
	user_id: string;
	space_id: string;
	name?: string;
	title: string;
	description: string;
	type: ContentType;
	source_url?: string;
	thumbnail_url?: string;
	created_at: string;
	updated_at: string;
};

export type BackendSpace = {
	id: string;
	user_id: string;
	name: string;
	description: string;
	created_at: string;
	updated_at: string;
};

export type BackendTag = {
	id: string;
	user_id: string;
	name: string;
	created_at: string;
};

export type ContentWithTags = {
	content: BackendContent;
	tags: BackendTag[];
};

export type IngestURLRequest = {
  user_id: string;
  space_id: string;
  name?: string;
  url: string;
};

export type IngestResult = {
	content_id: string;
	ingestion_id: string;
	status: "pending" | "processing" | "completed" | "failed";
	title: string;
	content_type: ContentType;
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
	content_ids?: string[];
	content_type?: ContentType;
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
	content_type: ContentType;
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

// =========================================================================
// PAUSED FEATURE: RAG Types (Put on hold for upcoming release)
// =========================================================================
// export type RAGScope =
// 	| { type: "library" }
// 	| { type: "content"; content_id: string }
// 	| { type: "selected_sources"; content_ids: string[] };
// 
// export type RAGMessage = {
// 	role: "user" | "assistant";
// 	content: string;
// };
// 
// export type RAGCitation = {
// 	content_id: string;
// 	title: string;
// 	type: ContentType;
// 	source_url?: string;
// 	page?: number;
// 	timestamp?: {
// 		start: number;
// 		end?: number;
// 	};
// };
// 
// export type RAGResponse = {
// 	answer: string;
// 	citations: RAGCitation[];
// };
// 
// export type RAGRequest = {
// 	user_id: string;
// 	question: string;
// 	scope: RAGScope;
// 	history?: RAGMessage[];
// 	top_k?: number;
// };

type APIResponse<T> = {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
};

async function apiJSON<T>(path: string, init?: RequestInit): Promise<T> {
	const headers: HeadersInit =
		init?.body instanceof FormData
			? init?.headers ?? {}
			: {
					"Content-Type": "application/json",
					...(init?.headers as Record<string, string> | undefined),
				};

	const response = await fetch(`${API_BASE_URL}${path}`, {
		...init,
		headers,
	});

	const body = (await response.json()) as APIResponse<T>;
	if (!response.ok || !body.success || body.data === undefined) {
		throw new Error(body.error || body.message || "Mindshelf request failed");
	}

	return body.data;
}

export async function listContent(userId?: string): Promise<BackendContent[]> {
	const qs = userId ? `?user_id=${encodeURIComponent(userId)}` : "";
	return apiJSON<BackendContent[]>(`/content${qs}`, {
		cache: "no-store",
	});
}

export async function getContent(id: string): Promise<BackendContent> {
	return apiJSON<BackendContent>(`/content/${id}`, {
		cache: "no-store",
	});
}

export async function updateContent(
	id: string,
	payload: {
		user_id: string;
		space_id: string;
		name: string;
		title: string;
		description: string;
		type: ContentType;
		source_url?: string;
		thumbnail_url?: string;
	}
): Promise<BackendContent> {
	return apiJSON<BackendContent>(`/content/${encodeURIComponent(id)}`, {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export async function deleteContent(id: string): Promise<void> {
	await apiJSON<unknown>(`/content/${encodeURIComponent(id)}`, {
		method: "DELETE",
	});
}

export async function listSpaces(userId?: string): Promise<BackendSpace[]> {
	const qs = userId ? `?user_id=${encodeURIComponent(userId)}` : "";
	return apiJSON<BackendSpace[]>(`/spaces${qs}`, {
		cache: "no-store",
	});
}

export async function syncUser(payload: {
	id: string;
	name: string;
	email: string;
}): Promise<unknown> {
	return apiJSON<unknown>("/users/sync", {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export async function createSpace(payload: {
	user_id: string;
	name: string;
	description: string;
}): Promise<BackendSpace> {
	return apiJSON<BackendSpace>("/spaces", {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export async function updateSpace(
	id: string,
	payload: { user_id: string; name: string; description: string }
): Promise<BackendSpace> {
	return apiJSON<BackendSpace>(`/spaces/${encodeURIComponent(id)}`, {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export async function deleteSpace(id: string): Promise<void> {
	await apiJSON<unknown>(`/spaces/${encodeURIComponent(id)}`, {
		method: "DELETE",
	});
}

export async function listTags(userId?: string): Promise<BackendTag[]> {
	const qs = userId ? `?user_id=${encodeURIComponent(userId)}` : "";
	return apiJSON<BackendTag[]>(`/tags${qs}`, {
		cache: "no-store",
	});
}

export async function createTag(payload: {
	user_id: string;
	name: string;
}): Promise<BackendTag> {
	return apiJSON<BackendTag>("/tags", {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export async function listContentTags(contentID: string): Promise<BackendTag[]> {
	return apiJSON<BackendTag[]>(`/content/${contentID}/tags`, {
		cache: "no-store",
	});
}

export async function setContentTags(
	contentID: string,
	tagIDs: string[]
): Promise<ContentWithTags> {
	return apiJSON<ContentWithTags>(`/content/${contentID}/tags`, {
		method: "PUT",
		body: JSON.stringify({ tag_ids: tagIDs }),
	});
}

export async function getIngestionResult(id: string): Promise<IngestResult> {
	return apiJSON<IngestResult>(`/ingestion/${id}`, {
		cache: "no-store",
	});
}

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
	name?: string;
	file: File;
}): Promise<IngestResult> {
	const formData = new FormData();
	formData.append("user_id", payload.user_id);
	formData.append("space_id", payload.space_id);
	if (payload.name?.trim()) {
		formData.append("name", payload.name.trim());
	}
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

export async function ingestFileWithProgress(
	payload: {
		user_id: string;
		space_id: string;
		name?: string;
		file: File;
	},
	onProgress: (progress: number) => void
): Promise<IngestResult> {
	const formData = new FormData();
	formData.append("user_id", payload.user_id);
	formData.append("space_id", payload.space_id);
	if (payload.name?.trim()) {
		formData.append("name", payload.name.trim());
	}
	formData.append("file", payload.file);

	return new Promise((resolve, reject) => {
		const request = new XMLHttpRequest();
		request.open("POST", `${API_BASE_URL}/ingestion/file`);

		request.upload.onprogress = (event) => {
			if (event.lengthComputable) {
				onProgress(Math.round((event.loaded / event.total) * 100));
			}
		};

		request.onload = () => {
			try {
				const body = JSON.parse(request.responseText) as APIResponse<IngestResult>;
				if (
					request.status < 200 ||
					request.status >= 300 ||
					!body.success ||
					!body.data
				) {
					reject(new Error(body.error || body.message || "Ingestion failed"));
					return;
				}
				onProgress(100);
				resolve(body.data);
			} catch {
				reject(new Error("Ingestion failed"));
			}
		};

		request.onerror = () => reject(new Error("Ingestion failed"));
		request.send(formData);
	});
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
	return apiJSON<SearchResponse>("/search", {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export const searchMindshelf = searchMemora;

export type ContentSummary = {
	summary: string;
	bullets: string[];
	source: string;
};

export async function getContentSummary(contentId: string): Promise<ContentSummary> {
	return apiJSON<ContentSummary>(`/content/${encodeURIComponent(contentId)}/summary`);
}

// =========================================================================
// PAUSED FEATURE: askMemora (Put on hold for upcoming release)
// =========================================================================
// export async function askMemora(payload: RAGRequest): Promise<RAGResponse> {
// 	return apiJSON<RAGResponse>("/rag", {
// 		method: "POST",
// 		body: JSON.stringify(payload),
// 	});
// }

export function displayContentName(item: BackendContent): string {
	return item.name?.trim() || item.title;
}

// Why this file exists:
// The frontend should call one typed API helper instead of scattering fetch details
// through components. This mirrors the backend service boundary in TypeScript.
