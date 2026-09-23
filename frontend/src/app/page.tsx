"use client";

import {
  ArrowRight,
  BrainCircuit,
  FileText,
  Image,
  Layers3,
  Play,
  Search,
  Sparkles,
  UploadCloud,
  Video,
  type LucideIcon,
} from "lucide-react";
import { motion, useMotionValueEvent, useScroll, useTransform } from "framer-motion";
import Link from "next/link";
import { useState } from "react";

import { ThemeToggle } from "@/components/theme/theme-toggle";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const trustItems = ["PDF", "DOCX", "YouTube", "Hybrid search", "pgvector", "Tags", "Spaces"];

const featureRows: {
  title: string;
  body: string;
  icon: LucideIcon;
  items: string[];
}[] = [
  {
    title: "Save anything worth keeping.",
    body: "Capture YouTube links, PDFs, DOCX files, articles, and images through the same Memora workflow.",
    icon: UploadCloud,
    items: ["YouTube transcript", "System Design Notes.pdf", "RAG Architecture.docx"],
  },
  {
    title: "Search by meaning.",
    body: "Use semantic, keyword, or hybrid search to find the idea you remember, even when the exact words are gone.",
    icon: Search,
    items: ["Kafka failures", "embedding contract", "goroutine cancellation"],
  },
  {
    title: "Organize with Spaces and Tags.",
    body: "Group learning, projects, interviews, and ideas into collections that mirror how you think.",
    icon: Layers3,
    items: ["AI / ML", "System Design", "Projects"],
  },
];

const previewCards: {
  type: string;
  title: string;
  excerpt: string;
  icon: LucideIcon;
}[] = [
  {
    type: "YouTube",
    title: "Kafka Consumer Groups Explained",
    excerpt: "Timestamp 01:24 - partition ownership",
    icon: Video,
  },
  {
    type: "PDF",
    title: "Distributed Systems Notes.pdf",
    excerpt: "Page 32 - consumer failures",
    icon: FileText,
  },
  {
    type: "DOCX",
    title: "RAG Architecture Notes.docx",
    excerpt: "Chunking and hybrid retrieval",
    icon: FileText,
  },
  {
    type: "Image",
    title: "Kafka consumer diagram",
    excerpt: "Tags: systems, scaling",
    icon: Image,
  },
];

const pipeline = [
  {
    label: "Add",
    body: "Paste a link or drop a file.",
    icon: UploadCloud,
  },
  {
    label: "Process",
    body: "Extract text, transcript, and metadata.",
    icon: BrainCircuit,
  },
  {
    label: "Search",
    body: "Query semantic and keyword indexes.",
    icon: Search,
  },
  {
    label: "Recall",
    body: "Open the exact source, page, or timestamp.",
    icon: Play,
  },
];

const spaces = ["AI / ML", "System Design", "College", "Projects", "Career", "Ideas"];

const containerVariants = {
  hidden: {},
  show: {
    transition: {
      staggerChildren: 0.08,
    },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 18 },
  show: { opacity: 1, y: 0 },
};

export default function LandingPage() {
  return (
    <main className="memora-landing">
      <FloatingNavbar />
      <section className="landing-shell relative mx-auto flex min-h-[calc(100vh-80px)] max-w-7xl flex-col justify-center px-4 py-16 sm:px-6 lg:px-8">
        <AppAlignedAccent />
        <motion.div
          className="mx-auto max-w-5xl text-center"
          initial="hidden"
          animate="show"
          variants={containerVariants}
        >
          <motion.div
            className="landing-eyebrow mx-auto inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm"
            variants={itemVariants}
          >
            <Sparkles className="size-4" />
            Your second memory
          </motion.div>
          <motion.h1
            className="mt-8 text-6xl font-bold leading-[0.95] tracking-normal sm:text-7xl lg:text-8xl"
            variants={itemVariants}
          >
            Save everything.
            <span className="sparkle-text block">Recall it instantly.</span>
          </motion.h1>
          <motion.p
            className="mx-auto mt-7 max-w-3xl text-xl font-medium leading-8 sm:text-2xl"
            variants={itemVariants}
          >
            Memora turns videos, documents, articles, images, and notes into a
            searchable knowledge space.
          </motion.p>
          <motion.div
            className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row"
            variants={itemVariants}
          >
            <Link
              className={cn(
                buttonVariants({ variant: "default", size: "lg" }),
                "shimmer-button h-12 rounded-full px-6 text-base"
              )}
              href="/signup"
            >
              <span className="relative z-10 inline-flex items-center gap-2">
                Get Started Free
                <ArrowRight className="size-4" />
              </span>
            </Link>
            <Link
              className="landing-ghost-button inline-flex h-12 items-center rounded-full px-6 text-base font-medium transition-colors hover:bg-accent"
              href="/login"
            >
              Sign in
            </Link>
          </motion.div>
          <motion.div
            className="trust-marquee mx-auto mt-7"
            variants={itemVariants}
            aria-label="Supported formats and features"
          >
            <div className="trust-marquee-track">
              {[...trustItems, ...trustItems].map((item, index) => (
                <span className="landing-card rounded-full px-3 py-1.5 text-xs landing-muted" key={`${item}-${index}`}>
                  {item}
                </span>
              ))}
            </div>
          </motion.div>
        </motion.div>

        <motion.div
          className="mt-16"
          initial={{ opacity: 0, y: 28 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, delay: 0.2 }}
        >
          <ProductPreview />
        </motion.div>
      </section>

      <FeatureRows />
      <AnimatedPipeline />
      <SpacesSection />
      <ClosingCTA />
      <LandingFooter />
    </main>
  );
}

function FloatingNavbar() {
  const { scrollY } = useScroll();
  const [scrolled, setScrolled] = useState(false);

  useMotionValueEvent(scrollY, "change", (latest) => {
    setScrolled(latest > 90);
  });

  return (
    <motion.header
      className={cn(
        "sticky top-0 z-30 border-b transition-all duration-200",
        scrolled ? "landing-nav" : "floating-nav-transparent"
      )}
    >
      <nav className="landing-shell mx-auto flex h-20 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <Link className="flex items-center gap-3 font-semibold tracking-normal" href="/">
          <span className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-lg">
            M
          </span>
          <span className="text-lg">MEMORA</span>
        </Link>
        <div className="hidden items-center gap-8 text-sm font-medium landing-muted md:flex">
          <a className="transition-colors hover:text-foreground" href="#product">
            Product
          </a>
          <a className="transition-colors hover:text-foreground" href="#pricing">
            Pricing
          </a>
          <a className="transition-colors hover:text-foreground" href="#docs">
            Docs
          </a>
        </div>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <Link
            className="hidden rounded-full px-4 py-2 text-sm font-medium transition-colors hover:bg-accent sm:inline-flex"
            href="/login"
          >
            Sign in
          </Link>
          <Link className={cn(buttonVariants({ variant: "default" }), "rounded-full")} href="/signup">
            Get Started
          </Link>
        </div>
      </nav>
    </motion.header>
  );
}

function AppAlignedAccent() {
  const { scrollYProgress } = useScroll();
  const y = useTransform(scrollYProgress, [0, 0.35], [0, 110]);
  const rotateX = useTransform(scrollYProgress, [0, 0.35], [58, 24]);
  const rotateY = useTransform(scrollYProgress, [0, 0.35], [-24, 32]);
  const rotateZ = useTransform(scrollYProgress, [0, 0.35], [8, -14]);

  return (
    <motion.div
      className="app-orb-3d pointer-events-none absolute right-12 top-40 hidden size-44 lg:block"
      style={{ y, rotateX, rotateY, rotateZ }}
      aria-hidden="true"
    >
      <span />
      <span />
      <span />
    </motion.div>
  );
}

function ProductPreview() {
  return (
    <div className="relative mx-auto max-w-6xl animate-[memora-float_8s_ease-in-out_infinite]">
      <div className="landing-panel overflow-hidden rounded-[2rem]">
        <div className="flex items-center justify-between border-b px-5 py-4">
          <div className="flex items-center gap-3">
            <span className="flex size-9 items-center justify-center rounded-xl bg-primary font-semibold text-primary-foreground">
              M
            </span>
            <span className="font-semibold">MEMORA</span>
          </div>
          <div className="hidden h-10 w-[42%] items-center gap-3 rounded-full border bg-background px-4 text-sm landing-muted md:flex">
            <Search className="size-4" />
            Search your memory...
            <kbd className="ml-auto rounded-md border px-2 py-1 font-mono text-xs">
              Ctrl K
            </kbd>
          </div>
        </div>

        <div className="grid min-h-[540px] lg:grid-cols-[220px_1fr]">
          <aside className="hidden border-r bg-background/40 p-5 lg:block">
            {["Home", "Search", "Library"].map((item, index) => (
              <div
                className={cn(
                  "mb-2 rounded-xl px-4 py-3 text-sm",
                  index === 0 ? "bg-accent text-accent-foreground" : "landing-muted"
                )}
                key={item}
              >
                {item}
              </div>
            ))}
            <p className="mt-8 px-4 text-xs font-semibold uppercase landing-faint">
              Spaces
            </p>
            {["AI / ML", "System Design", "Projects", "Learning"].map((item) => (
              <div className="mt-3 rounded-lg px-4 py-2 text-sm landing-muted" key={item}>
                {item}
              </div>
            ))}
          </aside>

          <section className="p-5 sm:p-7">
            <div className="grid gap-5 xl:grid-cols-[1fr_320px]">
              <div>
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-sm font-medium text-primary">Home</p>
                    <h2 className="mt-2 text-3xl font-semibold tracking-normal">
                      Your memory, organized.
                    </h2>
                  </div>
                  <button className="rounded-full bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground">
                    Add Content
                  </button>
                </div>

                <div className="mt-6 grid gap-4 sm:grid-cols-2">
                  {previewCards.map((card) => (
                    <MemoryPreviewCard card={card} key={card.title} />
                  ))}
                </div>
              </div>

              <aside className="rounded-2xl border bg-background/60 p-4">
                <div className="rounded-xl border bg-card p-4">
                  <p className="text-xs font-semibold uppercase landing-faint">Search</p>
                  <p className="mt-3 text-sm">kafka consumer failures</p>
                </div>
                <p className="mt-5 text-sm font-medium">7 memories found</p>
                <div className="mt-3 space-y-3">
                  {["Kafka Consumer Groups Explained", "Distributed Systems Notes", "Event Driven Architecture"].map((item) => (
                    <div className="landing-card rounded-xl p-3" key={item}>
                      <p className="text-sm font-semibold">{item}</p>
                      <p className="mt-1 text-xs landing-muted">
                        Relevant chunk with source metadata
                      </p>
                    </div>
                  ))}
                </div>
              </aside>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}

function MemoryPreviewCard({ card }: { card: (typeof previewCards)[number] }) {
  const Icon = card.icon;

  return (
    <motion.article
      className="landing-card overflow-hidden rounded-2xl"
      whileHover={{ y: -4, scale: 1.01 }}
      transition={{ duration: 0.18 }}
    >
      <div className="flex h-36 items-center justify-center bg-accent/40">
        <Icon className="size-10 text-primary" />
      </div>
      <div className="p-4">
        <p className="text-xs font-semibold uppercase landing-faint">{card.type}</p>
        <h3 className="mt-2 line-clamp-2 text-base font-semibold leading-6">
          {card.title}
        </h3>
        <p className="mt-2 text-sm landing-muted">{card.excerpt}</p>
      </div>
    </motion.article>
  );
}

function FeatureRows() {
  return (
    <section className="neon-section landing-shell" id="product">
      <div className="mx-auto max-w-7xl space-y-20 px-4 py-24 sm:px-6 lg:px-8">
        {featureRows.map((feature, index) => {
          const Icon = feature.icon;
          return (
            <motion.article
              className="grid gap-10 lg:grid-cols-2 lg:items-center"
              initial={{ opacity: 0, y: 24 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-120px" }}
              transition={{ duration: 0.3 }}
              key={feature.title}
            >
              <div className={cn("space-y-4", index % 2 === 1 && "lg:order-2")}>
                <p className="text-sm font-semibold uppercase text-primary">Feature</p>
                <h2 className="text-4xl font-semibold tracking-normal sm:text-5xl">
                  {feature.title}
                </h2>
                <p className="max-w-xl text-base leading-8 landing-muted">{feature.body}</p>
              </div>
              <div className="landing-panel rounded-[2rem] p-5">
                <div className="rounded-2xl border bg-background p-5">
                  <div className="flex size-12 items-center justify-center rounded-2xl bg-primary/15 text-primary">
                    <Icon className="size-6" />
                  </div>
                  <div className="mt-6 space-y-3">
                    {feature.items.map((item) => (
                      <div className="landing-card rounded-xl px-4 py-3 text-sm" key={item}>
                        {item}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </motion.article>
          );
        })}
      </div>
    </section>
  );
}

function AnimatedPipeline() {
  return (
    <section className="landing-shell mx-auto max-w-7xl px-4 py-24 sm:px-6 lg:px-8" id="docs">
      <div className="max-w-3xl">
        <p className="text-sm font-semibold uppercase text-primary">How it works</p>
        <h2 className="mt-4 text-4xl font-semibold tracking-normal sm:text-6xl">
          Add. Process. Search. Recall.
        </h2>
      </div>
      <div className="pipeline-shell mt-12">
        <div className="pipeline-line hidden md:block" />
        <div className="pipeline-orb hidden md:block" />
        <motion.div
          className="grid gap-4 md:grid-cols-4"
          initial="hidden"
          whileInView="show"
          viewport={{ once: true }}
          variants={containerVariants}
        >
          {pipeline.map((step) => {
            const Icon = step.icon;
            return (
              <motion.article className="landing-card relative rounded-2xl p-6" key={step.label} variants={itemVariants}>
                <div className="relative z-10 flex size-14 items-center justify-center rounded-2xl border bg-background text-primary">
                  <Icon className="size-6" />
                </div>
                <h3 className="mt-8 text-xl font-semibold">{step.label}</h3>
                <p className="mt-3 text-sm leading-7 landing-muted">{step.body}</p>
              </motion.article>
            );
          })}
        </motion.div>
      </div>
    </section>
  );
}

function SpacesSection() {
  return (
    <section className="neon-section landing-shell" id="pricing">
      <div className="mx-auto max-w-7xl px-4 py-24 sm:px-6 lg:px-8">
        <div className="max-w-3xl">
          <p className="text-sm font-semibold uppercase text-primary">Spaces</p>
          <h2 className="mt-4 text-4xl font-semibold tracking-normal sm:text-6xl">
            Organize knowledge around what matters to you.
          </h2>
        </div>
        <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {spaces.map((space) => (
            <motion.div
              className="landing-card rounded-2xl p-6"
              whileHover={{ y: -4 }}
              transition={{ duration: 0.18 }}
              key={space}
            >
              <Layers3 className="size-6 text-primary" />
              <p className="mt-5 text-lg font-semibold">{space}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

function ClosingCTA() {
  return (
    <section className="landing-shell relative mx-auto max-w-5xl px-4 py-28 text-center sm:px-6 lg:px-8">
      <div className="mx-auto flex size-14 items-center justify-center rounded-2xl border bg-card">
        <BrainCircuit className="size-7 text-primary" />
      </div>
      <h2 className="mt-8 text-5xl font-semibold tracking-normal sm:text-7xl">
        Remember more. Search less.
      </h2>
      <p className="mx-auto mt-6 max-w-2xl text-base leading-8 landing-muted">
        Build a personal knowledge space that preserves transcripts, extracted
        text, metadata, tags, and source context.
      </p>
      <Link
        className={cn(buttonVariants({ variant: "default", size: "lg" }), "shimmer-button mt-9 h-12 rounded-full px-6 text-base")}
        href="/signup"
      >
        <span className="relative z-10">Get Started Free</span>
      </Link>
    </section>
  );
}

function LandingFooter() {
  return (
    <footer className="neon-section landing-shell" id="about">
      <div className="mx-auto grid max-w-7xl gap-8 px-4 py-12 sm:px-6 md:grid-cols-[1fr_auto_auto_auto] lg:px-8">
        <div>
          <div className="flex items-center gap-3 font-semibold">
            <span className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground">
              M
            </span>
            MEMORA
          </div>
          <p className="mt-3 text-sm landing-muted">Your second memory.</p>
          <div className="mt-4">
            <ThemeToggle />
          </div>
        </div>
        <FooterGroup title="Product" items={["Features", "Pricing", "Security"]} />
        <FooterGroup title="Support" items={["Docs", "Status", "Contact"]} />
        <FooterGroup title="Legal" items={["Privacy", "Terms"]} />
      </div>
    </footer>
  );
}

function FooterGroup({ title, items }: { title: string; items: string[] }) {
  return (
    <div>
      <h3 className="text-sm font-semibold">{title}</h3>
      <div className="mt-3 space-y-2">
        {items.map((item) => (
          <p className="text-sm landing-muted" key={item}>
            {item}
          </p>
        ))}
      </div>
    </div>
  );
}
