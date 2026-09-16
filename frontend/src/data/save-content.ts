import { FileUp, ImageUp, Newspaper, PlaySquare, Rows3 } from "lucide-react";

import type { SaveInputOption } from "@/types/save-content";

export const saveInputOptions: SaveInputOption[] = [
  {
    value: "youtube",
    label: "YouTube",
    description: "Videos and Shorts with future transcript extraction.",
    icon: PlaySquare,
  },
  {
    value: "article",
    label: "Article",
    description: "Web pages and long-form references.",
    icon: Newspaper,
  },
  {
    value: "reddit",
    label: "Reddit",
    description: "Posts and discussions where accessible.",
    icon: Rows3,
  },
  {
    value: "file",
    label: "File",
    description: "PDF, DOCX, PPTX, TXT, CSV, and XLSX.",
    icon: FileUp,
  },
  {
    value: "image",
    label: "Image",
    description: "Images with future OCR and visual understanding.",
    icon: ImageUp,
  },
];

export const suggestedTags = ["RAG", "AI", "Go", "Kafka", "DSA", "VectorDB"];

export const supportedFileFormats = [
  "PDF",
  "DOCX",
  "PPTX",
  "TXT",
  "CSV",
  "XLSX",
  "Images",
];
