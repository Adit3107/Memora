import Link from "next/link";

type BrandProps = {
  appHref?: boolean;
  collapsed?: boolean;
};

export function Brand({ appHref = false, collapsed = false }: BrandProps) {
  const href = appHref ? "/" : "/";

  if (collapsed) {
    return (
      <Link
        className="flex items-center justify-center rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
        href={href}
        title="Mindshelf"
      >
        <span className="memora-wordmark text-xl leading-none">M</span>
      </Link>
    );
  }

  return (
    <Link
      className="flex items-center gap-3 rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={href}
    >
      <div className="min-w-0">
        <p className="memora-wordmark text-base leading-5 tracking-[-0.05em] sm:text-lg">
          Mindshelf
        </p>
        <p className="text-xs text-muted-foreground">Your intelligent shelf.</p>
      </div>
    </Link>
  );
}
