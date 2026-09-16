import { AlertTriangle } from "lucide-react";

import { Button } from "@/components/ui/button";

type ErrorStateProps = {
  title?: string;
  description?: string;
  onRetry?: () => void;
};

export function ErrorState({
  title = "Something went wrong",
  description = "The frontend route could not render correctly.",
  onRetry,
}: ErrorStateProps) {
  return (
    <div className="rounded-md border bg-card p-6 shadow-sm">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start">
        <div className="flex size-10 items-center justify-center rounded-md border bg-background">
          <AlertTriangle className="size-5 text-muted-foreground" aria-hidden="true" />
        </div>
        <div className="max-w-xl">
          <h1 className="text-xl font-semibold">{title}</h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            {description}
          </p>
          {onRetry ? (
            <Button className="mt-4" onClick={onRetry} variant="outline">
              Try again
            </Button>
          ) : null}
        </div>
      </div>
    </div>
  );
}
