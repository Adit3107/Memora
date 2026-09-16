type TagListProps = {
  tags: string[];
};

export function TagList({ tags }: TagListProps) {
  if (tags.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-2">
      {tags.map((tag) => (
        <span
          className="rounded-md border px-2 py-1 text-xs font-medium text-muted-foreground"
          key={tag}
        >
          #{tag}
        </span>
      ))}
    </div>
  );
}
