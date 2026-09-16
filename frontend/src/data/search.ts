import { savedContent } from "./content";

export const defaultSearchQuery =
  "Where did I save the video explaining vector databases?";

export const mockSearchSnippets: Record<string, string> = {
  "vector-databases-explained":
    "01:42 - Vector databases retrieve nearby ideas rather than exact words.",
  "rag-notes":
    "Page 31 - Evaluation checks answer faithfulness and retrieval quality.",
  "understanding-pgvector":
    "pgvector adds vector similarity search to PostgreSQL.",
  "kafka-consumer-group-diagram":
    "OCR placeholder mentions partitions, consumers, and offsets.",
  "go-pipelines":
    "Pipelines compose stages connected by channels.",
  "dsa-revision-sheet":
    "Sheet 3 - Graph traversal and shortest path notes.",
};

export const mockRelevance: Record<string, string> = {
  "vector-databases-explained": "High mock relevance",
  "rag-notes": "Medium mock relevance",
  "understanding-pgvector": "Medium mock relevance",
  "kafka-consumer-group-diagram": "Low mock relevance",
  "go-pipelines": "Low mock relevance",
  "dsa-revision-sheet": "Low mock relevance",
};

export const allTags = Array.from(
  new Set(savedContent.flatMap((item) => item.tags))
).sort((a, b) => a.localeCompare(b));
