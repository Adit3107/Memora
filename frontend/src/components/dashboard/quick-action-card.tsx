import Link from "next/link";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { QuickAction } from "@/types/dashboard";

type QuickActionCardProps = {
  action: QuickAction;
};

export function QuickActionCard({ action }: QuickActionCardProps) {
  const Icon = action.icon;

  return (
    <article className="flex min-h-40 flex-col justify-between rounded-md border bg-card p-5 shadow-sm">
      <div className="space-y-3">
        <div className="flex size-9 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <Icon className="size-4" aria-hidden="true" />
        </div>
        <div>
          <h2 className="text-base font-semibold">{action.title}</h2>
          <p className="mt-1 text-sm leading-6 text-muted-foreground">
            {action.description}
          </p>
        </div>
      </div>

      {action.href ? (
        <Link
          className={cn(buttonVariants({ variant: "secondary" }), "mt-5 w-fit")}
          href={action.href}
        >
          Open
        </Link>
      ) : (
        <button
          className={cn(
            buttonVariants({ variant: "outline" }),
            "mt-5 w-fit cursor-not-allowed opacity-60"
          )}
          disabled
          type="button"
        >
          Phase 1 placeholder
        </button>
      )}
    </article>
  );
}
