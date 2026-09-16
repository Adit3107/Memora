import { FileText, Image, Newspaper, Video } from "lucide-react";

import type { ContentType, SavedContentItem, Space } from "@/types/content";

export const spaces: Space[] = [
  {
    slug: "ai-learning",
    name: "AI Learning",
    description: "RAG, vector databases, embeddings, LLM notes, and AI papers.",
    ownerLabel: "Personal",
    updatedAt: "Updated today",
    tags: ["RAG", "AI", "VectorDB"],
  },
  {
    slug: "dsa",
    name: "DSA",
    description: "Algorithms, patterns, revision sheets, and interview prep.",
    ownerLabel: "Study",
    updatedAt: "Updated this week",
    tags: ["DSA", "Graphs", "DP"],
  },
  {
    slug: "go",
    name: "Go",
    description: "Go language references, concurrency examples, and backend notes.",
    ownerLabel: "Engineering",
    updatedAt: "Updated this week",
    tags: ["Go", "Backend", "Concurrency"],
  },
  {
    slug: "system-design",
    name: "System Design",
    description: "Architecture diagrams, distributed systems, Kafka, and scaling notes.",
    ownerLabel: "Engineering",
    updatedAt: "Updated yesterday",
    tags: ["Kafka", "Scaling", "Databases"],
  },
  {
    slug: "college",
    name: "College",
    description: "Class material and references waiting to be added.",
    ownerLabel: "Study",
    updatedAt: "No saved content yet",
    tags: ["Semester", "Notes"],
  },
];

export const contentTypeLabels: Record<ContentType, string> = {
  video: "Video",
  document: "Document",
  article: "Article",
  image: "Image",
};

export const savedContent: SavedContentItem[] = [
  {
    slug: "vector-databases-explained",
    title: "Vector Databases Explained",
    type: "video",
    source: "YouTube Short",
    sourceUrl: "https://youtube.com",
    description: "A quick explanation of embeddings, vector search, and similarity.",
    metadata: "Transcript ready mock - 04:18",
    dateLabel: "Saved today",
    spaceSlug: "ai-learning",
    spaceName: "AI Learning",
    tags: ["RAG", "VectorDB", "Embeddings"],
    icon: Video,
    detail: {
      heroLabel: "Video player placeholder",
      previewTitle: "Timestamped transcript preview",
      previewBody:
        "Future video ingestion will show the player or thumbnail here with transcript navigation.",
      extractedTitle: "Transcript placeholder",
      extractedBody: [
        "00:12 - Embeddings convert text into searchable numeric meaning.",
        "01:42 - Vector databases retrieve nearby ideas rather than exact words.",
        "03:10 - Hybrid search combines semantic and keyword matching.",
      ],
      referenceLabel: "Timestamps",
      references: ["00:12 Embeddings", "01:42 Vector search", "03:10 Hybrid search"],
    },
  },
  {
    slug: "rag-notes",
    title: "RAG Notes",
    type: "document",
    source: "PDF",
    description: "Study notes about retrieval augmented generation architecture.",
    metadata: "PDF - 42 pages",
    dateLabel: "Saved yesterday",
    spaceSlug: "ai-learning",
    spaceName: "AI Learning",
    tags: ["RAG", "LLM", "Citations"],
    icon: FileText,
    detail: {
      heroLabel: "Document preview placeholder",
      previewTitle: "File preview",
      previewBody:
        "Future document processing will render file previews and extracted text sections.",
      extractedTitle: "Extracted text placeholder",
      extractedBody: [
        "Page 3 - RAG systems retrieve relevant chunks before generation.",
        "Page 18 - Citations should point back to source chunks.",
        "Page 31 - Evaluation checks answer faithfulness and retrieval quality.",
      ],
      referenceLabel: "Pages",
      references: ["Page 3", "Page 18", "Page 31"],
    },
  },
  {
    slug: "understanding-pgvector",
    title: "Understanding pgvector",
    type: "article",
    source: "Web article",
    sourceUrl: "https://example.com/pgvector",
    description: "Article notes on storing and querying embeddings in PostgreSQL.",
    metadata: "Article - 9 min read",
    dateLabel: "Saved today",
    spaceSlug: "ai-learning",
    spaceName: "AI Learning",
    tags: ["PostgreSQL", "VectorDB", "Search"],
    icon: Newspaper,
    detail: {
      heroLabel: "Article preview placeholder",
      previewTitle: "Readable article view",
      previewBody:
        "Future article ingestion will display cleaned text, source metadata, and crawl details.",
      extractedTitle: "Article content placeholder",
      extractedBody: [
        "pgvector adds vector similarity search to PostgreSQL.",
        "Indexes help approximate nearest-neighbor search at larger scale.",
        "Metadata filters can narrow retrieval before ranking.",
      ],
      referenceLabel: "Source sections",
      references: ["Intro", "Indexes", "Filtering"],
    },
  },
  {
    slug: "kafka-consumer-group-diagram",
    title: "Kafka Consumer Group Diagram",
    type: "image",
    source: "Image upload",
    description: "Visual reference for partitions, consumers, and group coordination.",
    metadata: "PNG - diagram",
    dateLabel: "Saved yesterday",
    spaceSlug: "system-design",
    spaceName: "System Design",
    tags: ["Kafka", "Architecture", "Diagram"],
    icon: Image,
    detail: {
      heroLabel: "Image preview placeholder",
      previewTitle: "Image metadata",
      previewBody:
        "Future image understanding will show previews, OCR text, and visual summaries.",
      extractedTitle: "OCR text placeholder",
      extractedBody: [
        "Consumer Group A reads partitions 0, 1, and 2.",
        "Rebalancing assigns partitions when consumers join or leave.",
        "Offsets track progress per consumer group.",
      ],
      referenceLabel: "Image notes",
      references: ["Partitions", "Consumers", "Offsets"],
    },
  },
  {
    slug: "go-pipelines",
    title: "Go Pipelines and Cancellation",
    type: "article",
    source: "Go blog",
    sourceUrl: "https://go.dev/blog/pipelines",
    description: "Notes on channel pipelines, cancellation, and goroutine cleanup.",
    metadata: "Article - 12 min read",
    dateLabel: "Saved this week",
    spaceSlug: "go",
    spaceName: "Go",
    tags: ["Go", "Concurrency", "Backend"],
    icon: Newspaper,
    detail: {
      heroLabel: "Article preview placeholder",
      previewTitle: "Technical article view",
      previewBody:
        "Future article ingestion can preserve headings, code snippets, and source links.",
      extractedTitle: "Article content placeholder",
      extractedBody: [
        "Pipelines compose stages connected by channels.",
        "Cancellation prevents blocked sends and goroutine leaks.",
        "Context propagation keeps long-running work manageable.",
      ],
      referenceLabel: "Sections",
      references: ["Pipelines", "Cancellation", "Context"],
    },
  },
  {
    slug: "dsa-revision-sheet",
    title: "DSA Revision Sheet",
    type: "document",
    source: "Spreadsheet",
    description: "Pattern-based revision tracker for arrays, graphs, and DP.",
    metadata: "XLSX - 6 sheets",
    dateLabel: "Saved this week",
    spaceSlug: "dsa",
    spaceName: "DSA",
    tags: ["DSA", "Revision", "Graphs"],
    icon: FileText,
    detail: {
      heroLabel: "Spreadsheet preview placeholder",
      previewTitle: "Tabular preview",
      previewBody:
        "Future spreadsheet handling will expose sheet summaries and extracted rows.",
      extractedTitle: "Extracted table placeholder",
      extractedBody: [
        "Sheet 1 - Array patterns and two-pointer problems.",
        "Sheet 3 - Graph traversal and shortest path notes.",
        "Sheet 5 - Dynamic programming recurrence examples.",
      ],
      referenceLabel: "Sheets",
      references: ["Arrays", "Graphs", "DP"],
    },
  },
];

export function getSpaceBySlug(slug: string) {
  return spaces.find((space) => space.slug === slug);
}

export function getContentBySlug(slug: string) {
  return savedContent.find((item) => item.slug === slug);
}

export function getContentBySpace(spaceSlug: string) {
  return savedContent.filter((item) => item.spaceSlug === spaceSlug);
}
