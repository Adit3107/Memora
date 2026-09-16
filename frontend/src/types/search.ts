import type { ContentType } from "./content";

export type SearchFilter = {
  contentType: "all" | ContentType;
  spaceSlug: "all" | string;
  tag: "all" | string;
};
