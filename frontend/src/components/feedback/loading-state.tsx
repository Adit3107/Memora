type LoadingStateProps = {
  title?: string;
  description?: string;
};

export function LoadingState({
  title = "Loading Mindshelf",
  description = "Preparing the interface.",
}: LoadingStateProps) {
  return (
    <div className="rounded-md border bg-card p-6 shadow-sm">
      <div className="space-y-4">
        <div className="h-4 w-36 animate-pulse rounded-md bg-muted" />
        <div className="h-8 w-full max-w-xl animate-pulse rounded-md bg-muted" />
        <div className="grid gap-3 md:grid-cols-3">
          <div className="h-24 animate-pulse rounded-md bg-muted" />
          <div className="h-24 animate-pulse rounded-md bg-muted" />
          <div className="h-24 animate-pulse rounded-md bg-muted" />
        </div>
      </div>
      <span className="sr-only">
        {title}. {description}
      </span>
    </div>
  );
}
