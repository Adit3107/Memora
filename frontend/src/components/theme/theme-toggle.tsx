"use client";

import { Monitor, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

type ThemeMode = "light" | "dark" | "system";

const modes: { label: string; value: ThemeMode; icon: typeof Sun }[] = [
  { label: "Dark", value: "dark", icon: Moon },
  { label: "Light", value: "light", icon: Sun },
  { label: "System", value: "system", icon: Monitor },
];

export function ThemeToggle() {
  const [mode, setMode] = useState<ThemeMode>("system");

  useEffect(() => {
    const saved = window.localStorage.getItem("memora-theme") as ThemeMode | null;
    const initial = saved ?? "dark";
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMode(initial);
    applyTheme(initial);
  }, []);

  function cycleTheme() {
    const currentIndex = modes.findIndex((item) => item.value === mode);
    const next = modes[(currentIndex + 1) % modes.length].value;
    setMode(next);
    window.localStorage.setItem("memora-theme", next);
    applyTheme(next);
  }

  const Icon = modes.find((item) => item.value === mode)?.icon ?? Monitor;

  return (
    <button
      aria-label={`Theme: ${mode}`}
      className="inline-flex size-9 items-center justify-center rounded-md border bg-card text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
      onClick={cycleTheme}
      title={`Theme: ${mode}`}
      type="button"
    >
      <Icon className="size-4" />
    </button>
  );
}

function applyTheme(mode: ThemeMode) {
  const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  const shouldUseDark = mode === "dark" || (mode === "system" && prefersDark);
  document.documentElement.classList.remove("dark", "light");
  document.documentElement.classList.add(shouldUseDark ? "dark" : "light");
}
