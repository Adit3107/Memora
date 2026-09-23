import { ArrowRight, Mail } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function LoginPage() {
  return (
    <AuthShell
      asideTitle="Your second memory."
      asideText="Save anything. Memora understands it. Find it when you need it."
    >
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold">Welcome back</h1>
        <p className="text-sm text-muted-foreground">Continue your memory.</p>
      </div>

      <form className="mt-8 space-y-4">
        <AuthInput label="Email" type="email" />
        <AuthInput label="Password" type="password" />
        <Link className={cn(buttonVariants({ variant: "default" }), "h-10 w-full")} href="/app">
          Sign in
          <ArrowRight className="size-4" />
        </Link>
      </form>

      <div className="mt-5 flex items-center justify-between text-sm">
        <Link className="text-muted-foreground hover:text-foreground" href="/forgot-password">
          Forgot password?
        </Link>
      </div>

      <div className="my-8 flex items-center gap-3 text-xs uppercase text-muted-foreground">
        <span className="h-px flex-1 bg-border" />
        OR
        <span className="h-px flex-1 bg-border" />
      </div>

      <Link className={cn(buttonVariants({ variant: "outline" }), "h-10 w-full")} href="/app">
        <Mail className="size-4" />
        Continue with Google
      </Link>

      <Link className={cn(buttonVariants({ variant: "secondary" }), "mt-3 h-10 w-full")} href="/app">
        Continue as Demo User
      </Link>

      <p className="mt-8 text-center text-sm text-muted-foreground">
        Do not have an account?{" "}
        <Link className="font-medium text-foreground hover:underline" href="/signup">
          Create one
        </Link>
      </p>
    </AuthShell>
  );
}

function AuthShell({
  asideText,
  asideTitle,
  children,
}: {
  asideText: string;
  asideTitle: string;
  children: ReactNode;
}) {
  return (
    <main className="memora-landing grid min-h-screen lg:grid-cols-[1.05fr_0.95fr]">
      <section className="landing-shell hidden border-r bg-card/70 p-10 backdrop-blur lg:flex lg:flex-col lg:justify-between">
        <Link className="flex items-center gap-3 font-semibold" href="/">
          <span className="flex size-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            M
          </span>
          MEMORA
        </Link>
        <div className="max-w-lg">
          <div className="mb-8 grid grid-cols-2 gap-3">
            {["Videos", "Docs", "Search", "Spaces"].map((item) => (
              <div className="rounded-lg border bg-background p-5 text-sm font-medium" key={item}>
                {item}
              </div>
            ))}
          </div>
          <h2 className="text-5xl font-semibold tracking-normal">{asideTitle}</h2>
          <p className="mt-5 text-base leading-8 text-muted-foreground">{asideText}</p>
        </div>
        <p className="text-sm text-muted-foreground">MEMORA - Your second memory.</p>
      </section>
      <section className="landing-shell flex items-center justify-center px-4 py-12">
        <div className="landing-panel w-full max-w-md rounded-2xl p-6 sm:p-8">
          {children}
        </div>
      </section>
    </main>
  );
}

function AuthInput({ label, type }: { label: string; type: string }) {
  return (
    <label className="space-y-2 text-sm font-medium">
      {label}
      <input
        className="h-11 w-full rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
        placeholder={label}
        type={type}
      />
    </label>
  );
}
