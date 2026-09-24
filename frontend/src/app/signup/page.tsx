"use client";

import { useSignUp } from "@clerk/nextjs/legacy";
import {
  AlertCircle,
  ArrowRight,
  CheckCircle2,
  Eye,
  EyeOff,
  KeyRound,
  Lock,
  Mail,
  RefreshCw,
  Sparkles,
  User,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

export default function SignupPage() {
  const { isLoaded, signUp, setActive } = useSignUp();
  const router = useRouter();

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [emailAddress, setEmailAddress] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  // Email OTP verification state
  const [pendingVerification, setPendingVerification] = useState(false);
  const [code, setCode] = useState("");

  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [resendSuccess, setResendSuccess] = useState(false);

  // Step 1: Create user & send email verification OTP
  async function handleSignUp(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!isLoaded) return;
    setError("");
    setLoading(true);

    try {
      await signUp.create({
        firstName: firstName.trim() || undefined,
        lastName: lastName.trim() || undefined,
        emailAddress: emailAddress.trim(),
        password,
      });

      // Send the email with the OTP verification code
      await signUp.prepareEmailAddressVerification({ strategy: "email_code" });
      setPendingVerification(true);
    } catch (err: unknown) {
      const clerkErr = err as { errors?: Array<{ message: string }> };
      setError(
        clerkErr.errors?.[0]?.message ||
          "Could not create account. Please check your details and try again."
      );
    } finally {
      setLoading(false);
    }
  }

  // Step 2: Verify email address using the OTP code sent to user's inbox
  async function handleVerify(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!isLoaded) return;
    setError("");
    setLoading(true);

    try {
      const completeSignUp = await signUp.attemptEmailAddressVerification({
        code: code.trim(),
      });

      if (completeSignUp.status === "complete") {
        await setActive({ session: completeSignUp.createdSessionId });
        router.push("/app");
      } else {
        setError("Verification incomplete. Please check your code or try resending.");
      }
    } catch (err: unknown) {
      const clerkErr = err as { errors?: Array<{ message: string }> };
      setError(
        clerkErr.errors?.[0]?.message ||
          "Invalid verification code. Please double-check and try again."
      );
    } finally {
      setLoading(false);
    }
  }

  // Resend OTP code
  async function handleResendCode() {
    if (!isLoaded || resending) return;
    setError("");
    setResending(true);
    setResendSuccess(false);

    try {
      await signUp.prepareEmailAddressVerification({ strategy: "email_code" });
      setResendSuccess(true);
      setTimeout(() => setResendSuccess(false), 5000);
    } catch (err: unknown) {
      const clerkErr = err as { errors?: Array<{ message: string }> };
      setError(clerkErr.errors?.[0]?.message || "Failed to resend code.");
    } finally {
      setResending(false);
    }
  }

  // Social sign up via Google
  function handleGoogleSignUp() {
    if (!isLoaded || !signUp) return;
    void signUp.authenticateWithRedirect({
      strategy: "oauth_google",
      redirectUrl: "/sso-callback",
      redirectUrlComplete: "/app",
    });
  }

  return (
    <main className="memora-landing relative min-h-screen overflow-hidden bg-background">
      {/* Decorative ambient background glows */}
      <div className="pointer-events-none absolute -left-40 -top-40 size-[500px] rounded-full bg-primary/10 blur-[130px]" />
      <div className="pointer-events-none absolute -bottom-40 -right-40 size-[500px] rounded-full bg-primary/15 blur-[140px]" />

      <div className="relative mx-auto grid min-h-screen max-w-7xl items-center p-4 sm:p-6 lg:grid-cols-[1.1fr_0.9fr] lg:p-12">
        {/* Left Column: Dribbble-inspired feature branding */}
        <section className="hidden flex-col justify-between py-12 pr-12 lg:flex">
          <div>
            <Link className="inline-flex items-center gap-3 font-semibold" href="/">
              <span className="memora-wordmark text-2xl tracking-tight text-foreground">
                Mindshelf
              </span>
              <span className="rounded-full border border-primary/30 bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary">
                v2.0
              </span>
            </Link>

            <div className="mt-16 max-w-lg space-y-5">
              <div className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-card/60 px-3.5 py-1 text-xs font-medium backdrop-blur">
                <Sparkles className="size-3.5 text-primary" />
                <span>Whisper + RapidOCR + Gemini AI</span>
              </div>
              <h1 className="font-display text-5xl font-bold leading-[1.15] tracking-tight text-foreground">
                Your intelligent shelf for videos, reels, and documents.
              </h1>
              <p className="text-base leading-relaxed text-muted-foreground">
                Turn fast-moving social content, YouTube masterclasses, and dense PDFs into an
                organized second memory with timestamped transcripts and AI summaries.
              </p>
            </div>
          </div>

          <div className="grid max-w-lg grid-cols-2 gap-3.5 pt-10">
            <div className="rounded-xl border border-border/80 bg-card/50 p-4 backdrop-blur-sm">
              <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Reel Capture
              </p>
              <p className="mt-1 text-sm font-semibold text-foreground">
                Whisper Audio + RapidOCR
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Extracts speech and visual text simultaneously.
              </p>
            </div>
            <div className="rounded-xl border border-border/80 bg-card/50 p-4 backdrop-blur-sm">
              <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                AI Summaries
              </p>
              <p className="mt-1 text-sm font-semibold text-foreground">
                Gemini 3.8 Flash
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Executive takeaways in under 2 seconds.
              </p>
            </div>
          </div>
        </section>

        {/* Right Column: Custom Auth Card */}
        <section className="flex w-full items-center justify-center">
          <div className="w-full max-w-md rounded-2xl border border-border/80 bg-card/80 p-6 shadow-2xl backdrop-blur-xl sm:p-8">
            <div className="mb-6 lg:hidden">
              <Link className="inline-flex items-center gap-2 font-semibold" href="/">
                <span className="memora-wordmark text-2xl text-foreground">Mindshelf</span>
              </Link>
            </div>

            {!pendingVerification ? (
              /* Step 1: Sign up form */
              <div>
                <div className="space-y-1">
                  <h2 className="text-2xl font-bold tracking-tight text-foreground">
                    Create your account
                  </h2>
                  <p className="text-sm text-muted-foreground">
                    Start organizing your digital knowledge shelf.
                  </p>
                </div>

                {error ? (
                  <div className="mt-4 flex items-start gap-2.5 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
                    <AlertCircle className="size-4 shrink-0" />
                    <span>{error}</span>
                  </div>
                ) : null}

                <form className="mt-6 space-y-4" onSubmit={handleSignUp}>
                  <div className="grid grid-cols-2 gap-3">
                    <label className="space-y-1.5 text-xs font-medium text-foreground">
                      First Name
                      <div className="relative">
                        <User className="absolute left-3 top-3 size-4 text-muted-foreground" />
                        <input
                          className="h-10 w-full rounded-lg border border-border bg-background/80 pl-9 pr-3 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
                          onChange={(e) => setFirstName(e.target.value)}
                          placeholder="Aditya"
                          type="text"
                          value={firstName}
                        />
                      </div>
                    </label>
                    <label className="space-y-1.5 text-xs font-medium text-foreground">
                      Last Name
                      <input
                        className="h-10 w-full rounded-lg border border-border bg-background/80 px-3 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
                        onChange={(e) => setLastName(e.target.value)}
                        placeholder="Sharma"
                        type="text"
                        value={lastName}
                      />
                    </label>
                  </div>

                  <label className="space-y-1.5 text-xs font-medium text-foreground">
                    Email address
                    <div className="relative">
                      <Mail className="absolute left-3 top-3 size-4 text-muted-foreground" />
                      <input
                        className="h-10 w-full rounded-lg border border-border bg-background/80 pl-9 pr-3 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
                        onChange={(e) => setEmailAddress(e.target.value)}
                        placeholder="you@domain.com"
                        required
                        type="email"
                        value={emailAddress}
                      />
                    </div>
                  </label>

                  <label className="space-y-1.5 text-xs font-medium text-foreground">
                    Password
                    <div className="relative">
                      <Lock className="absolute left-3 top-3 size-4 text-muted-foreground" />
                      <input
                        className="h-10 w-full rounded-lg border border-border bg-background/80 pl-9 pr-10 text-sm outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="Create a strong password"
                        required
                        type={showPassword ? "text" : "password"}
                        value={password}
                      />
                      <button
                        className="absolute right-3 top-3 text-muted-foreground hover:text-foreground"
                        onClick={() => setShowPassword(!showPassword)}
                        type="button"
                      >
                        {showPassword ? (
                          <EyeOff className="size-4" />
                        ) : (
                          <Eye className="size-4" />
                        )}
                      </button>
                    </div>
                  </label>

                  <button
                    className="mt-2 flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-primary font-medium text-primary-foreground shadow-md transition hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/30 disabled:opacity-50"
                    disabled={loading}
                    type="submit"
                  >
                    {loading ? (
                      <RefreshCw className="size-4 animate-spin" />
                    ) : (
                      <>
                        <span>Continue</span>
                        <ArrowRight className="size-4" />
                      </>
                    )}
                  </button>
                </form>

                <div className="my-6 flex items-center gap-3 text-xs uppercase text-muted-foreground">
                  <span className="h-px flex-1 bg-border" />
                  <span>or</span>
                  <span className="h-px flex-1 bg-border" />
                </div>

                <div className="space-y-2.5">
                  <button
                    className="flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-border bg-background/80 text-sm font-medium transition hover:bg-accent focus:outline-none focus:ring-2 focus:ring-primary/20"
                    onClick={handleGoogleSignUp}
                    type="button"
                  >
                    <svg className="size-4" viewBox="0 0 24 24">
                      <path
                        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                        fill="#4285F4"
                      />
                      <path
                        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                        fill="#34A853"
                      />
                      <path
                        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"
                        fill="#FBBC05"
                      />
                      <path
                        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
                        fill="#EA4335"
                      />
                    </svg>
                    <span>Continue with Google</span>
                  </button>

                  <Link
                    className="flex h-10 w-full items-center justify-center rounded-lg border border-border/70 bg-secondary/40 text-sm font-medium text-foreground transition hover:bg-secondary focus:outline-none focus:ring-2 focus:ring-primary/20"
                    href="/app"
                  >
                    Continue as Demo User
                  </Link>
                </div>

                <p className="mt-6 text-center text-xs text-muted-foreground">
                  Already have an account?{" "}
                  <Link
                    className="font-semibold text-primary underline-offset-4 hover:underline"
                    href="/login"
                  >
                    Sign in
                  </Link>
                </p>
              </div>
            ) : (
              /* Step 2: Email OTP Verification Card */
              <div>
                <div className="flex size-12 items-center justify-center rounded-full border border-primary/20 bg-primary/10 text-primary">
                  <KeyRound className="size-6" />
                </div>

                <div className="mt-4 space-y-1">
                  <h2 className="text-2xl font-bold tracking-tight text-foreground">
                    Verify your email
                  </h2>
                  <p className="text-sm text-muted-foreground">
                    We sent a 6-digit verification code to{" "}
                    <span className="font-semibold text-foreground">{emailAddress}</span>.
                  </p>
                </div>

                {error ? (
                  <div className="mt-4 flex items-start gap-2.5 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
                    <AlertCircle className="size-4 shrink-0" />
                    <span>{error}</span>
                  </div>
                ) : null}

                {resendSuccess ? (
                  <div className="mt-4 flex items-center gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs text-emerald-400">
                    <CheckCircle2 className="size-4 shrink-0" />
                    <span>A new 6-digit code has been sent to your email.</span>
                  </div>
                ) : null}

                <form className="mt-6 space-y-4" onSubmit={handleVerify}>
                  <label className="space-y-2 text-xs font-medium text-foreground">
                    Enter 6-digit Code
                    <input
                      autoComplete="one-time-code"
                      autoFocus
                      className="h-12 w-full rounded-lg border border-border bg-background/80 text-center font-mono text-xl tracking-[0.3em] outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/20"
                      maxLength={6}
                      onChange={(e) => setCode(e.target.value)}
                      placeholder="······"
                      required
                      type="text"
                      value={code}
                    />
                  </label>

                  <button
                    className="flex h-11 w-full items-center justify-center gap-2 rounded-lg bg-primary font-medium text-primary-foreground shadow-md transition hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary/30 disabled:opacity-50"
                    disabled={loading || code.trim().length < 6}
                    type="submit"
                  >
                    {loading ? (
                      <RefreshCw className="size-4 animate-spin" />
                    ) : (
                      <>
                        <span>Verify & Launch Mindshelf</span>
                        <ArrowRight className="size-4" />
                      </>
                    )}
                  </button>
                </form>

                <div className="mt-6 flex flex-col items-center justify-between gap-3 text-xs sm:flex-row">
                  <button
                    className="text-muted-foreground transition hover:text-foreground disabled:opacity-50"
                    disabled={resending}
                    onClick={handleResendCode}
                    type="button"
                  >
                    {resending ? "Sending new code..." : "Resend code"}
                  </button>
                  <button
                    className="text-muted-foreground transition hover:text-foreground"
                    onClick={() => {
                      setPendingVerification(false);
                      setError("");
                    }}
                    type="button"
                  >
                    Change email address
                  </button>
                </div>
              </div>
            )}
          </div>
        </section>
      </div>
    </main>
  );
}
