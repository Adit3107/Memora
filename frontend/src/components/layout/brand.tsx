import Link from "next/link";

type BrandProps = {
  appHref?: boolean;
  collapsed?: boolean;
};

export function Brand({ appHref = false, collapsed = false }: BrandProps) {
  const href = appHref ? "/" : "/";

  return (
    <Link
      className="flex items-center gap-3 rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={href}
    >
      <div className="min-w-0">
        <p className="memora-wordmark text-base leading-5 tracking-[-0.08em] sm:text-lg">
          Memora
        </p>
        {!collapsed ? (
          <p className="text-xs text-muted-foreground">Your second memory.</p>
        ) : null}
      </div>
    </Link>
  );
}
