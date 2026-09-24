import Link from "next/link";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function ForgotPasswordPage() {
  return (
    <main className="memora-landing flex min-h-screen items-center justify-center px-4 py-12">
      <section className="landing-panel landing-shell w-full max-w-md rounded-2xl p-8">
        <Link className="flex items-center gap-3 font-semibold" href="/">
          <span className="memora-wordmark text-2xl">Mindshelf</span>
        </Link>
        <h1 className="mt-10 text-3xl font-semibold">Reset your password</h1>
        <p className="mt-3 text-sm leading-6 text-muted-foreground">
          Enter your email and Mindshelf will send a password reset code to your inbox.
        </p>
        <form className="mt-8 space-y-4">
          <label className="space-y-2 text-sm font-medium">
            Email
            <input
              className="h-11 w-full rounded-md border bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
              placeholder="aditya@example.com"
              type="email"
            />
          </label>
          <Link className={cn(buttonVariants({ variant: "default" }), "h-10 w-full")} href="/login">
            Send reset link
          </Link>
          <Link className={cn(buttonVariants({ variant: "secondary" }), "h-10 w-full")} href="/app">
            Continue as Demo User
          </Link>
        </form>
        <Link className="mt-6 block text-center text-sm text-muted-foreground hover:text-foreground" href="/login">
          Back to sign in
        </Link>
      </section>
    </main>
  );
}
