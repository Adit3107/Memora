import { ArrowRight } from "lucide-react";
import Link from "next/link";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function SignupPage() {
  return (
    <main className="memora-landing flex min-h-screen items-center justify-center px-4 py-12">
      <section className="landing-panel landing-shell grid w-full max-w-5xl overflow-hidden rounded-2xl lg:grid-cols-[0.95fr_1.05fr]">
        <div className="border-b p-8 lg:border-b-0 lg:border-r">
          <Link className="flex items-center gap-3 font-semibold" href="/">
            <span className="memora-wordmark text-2xl">Memora</span>
          </Link>
          <h1 className="mt-14 text-4xl font-semibold tracking-normal">
            Create your Memora
          </h1>
          <p className="mt-4 text-base leading-7 text-muted-foreground">
            Start building your second memory with a workspace ready for real
            ingestion, library browsing, and search.
          </p>
        </div>
        <div className="p-8">
          <form className="space-y-4">
            <AuthInput label="Name" type="text" />
            <AuthInput label="Email" type="email" />
            <AuthInput label="Password" type="password" />
            <Link className={cn(buttonVariants({ variant: "default" }), "h-10 w-full")} href="/app">
              Create account
              <ArrowRight className="size-4" />
            </Link>
            <Link className={cn(buttonVariants({ variant: "secondary" }), "h-10 w-full")} href="/app">
              Continue as Demo User
            </Link>
          </form>
          <p className="mt-8 text-center text-sm text-muted-foreground">
            Already have an account?{" "}
            <Link className="font-medium text-foreground hover:underline" href="/login">
              Sign in
            </Link>
          </p>
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
