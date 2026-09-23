import Link from "next/link";

type BrandProps = {
  appHref?: boolean;
  collapsed?: boolean;
};

export function Brand({ appHref = false, collapsed = false }: BrandProps) {
  return (
    <Link
      className="flex items-center gap-3 rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      href={appHref ? "/app" : "/"}
    >
      <div className="flex size-9 items-center justify-center rounded-lg bg-primary text-sm font-semibold text-primary-foreground">
        M
      </div>
      {!collapsed ? (
        <div className="min-w-0">
          <p className="text-base font-semibold leading-5">Memora</p>
          <p className="text-xs text-muted-foreground">Your second memory.</p>
        </div>
      ) : null}
    </Link>
  );
}
