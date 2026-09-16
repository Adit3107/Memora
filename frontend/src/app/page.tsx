import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <main className="min-h-screen bg-background text-foreground">
      <section className="mx-auto flex min-h-screen w-full max-w-6xl flex-col px-6 py-8">
        <header className="flex items-center justify-between border-b pb-5">
          <div>
            <p className="text-sm font-medium text-muted-foreground">
              Your second memory.
            </p>
            <h1 className="text-2xl font-semibold tracking-normal">Memora</h1>
          </div>
          <Button variant="outline">Phase 1</Button>
        </header>

        <div className="grid flex-1 items-center gap-10 py-12 lg:grid-cols-[1.05fr_0.95fr]">
          <div className="space-y-6">
            <div className="space-y-4">
              <p className="text-sm font-medium uppercase tracking-normal text-muted-foreground">
                Frontend foundation
              </p>
              <h2 className="max-w-3xl text-4xl font-semibold tracking-normal sm:text-5xl">
                Save anything. Find it when your brain taps out.
              </h2>
              <p className="max-w-2xl text-lg leading-8 text-muted-foreground">
                Memora will become a multimodal knowledge base for videos,
                documents, articles, images, and notes. This phase establishes
                the UI shell without connecting to backend, storage, AI, or
                ingestion services.
              </p>
            </div>

            <div className="flex flex-wrap gap-3">
              <Button>Open dashboard</Button>
              <Button variant="secondary">Save content</Button>
            </div>
          </div>

          <div className="rounded-lg border bg-card p-5 text-card-foreground shadow-sm">
            <div className="border-b pb-4">
              <p className="text-sm font-medium text-muted-foreground">
                Phase 1 modules
              </p>
            </div>
            <div className="mt-5 grid gap-3">
              {[
                "Global layout and navigation",
                "Dashboard",
                "Spaces UI",
                "Content library",
                "Content detail UI",
                "Upload and save content UI",
                "Search UI",
              ].map((item) => (
                <div
                  className="flex items-center justify-between rounded-md border px-4 py-3"
                  key={item}
                >
                  <span className="text-sm font-medium">{item}</span>
                  <span className="text-xs text-muted-foreground">Ready</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
