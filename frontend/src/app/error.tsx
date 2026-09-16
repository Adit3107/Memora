"use client";

import { ErrorState } from "@/components/feedback/error-state";

type ErrorPageProps = {
  reset: () => void;
};

export default function ErrorPage({ reset }: ErrorPageProps) {
  return (
    <ErrorState
      description="This is a frontend error boundary for Phase 1 routes."
      onRetry={reset}
    />
  );
}
