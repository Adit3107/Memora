"use client";

import { useUser } from "@clerk/nextjs";
import { useEffect, useState } from "react";

export function UserGreeting() {
  const { user, isLoaded } = useUser();
  const [greeting, setGreeting] = useState("Welcome");

  useEffect(() => {
    setGreeting(greetingForHour(new Date().getHours()));
  }, []);

  const name =
    user?.firstName ||
    (user?.fullName ? user.fullName.split(" ")[0] : null) ||
    "there";

  return (
    <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl min-h-[2.5rem]">
      {isLoaded && user ? `${greeting}, ${name}.` : `${greeting}.`}
    </h1>
  );
}

function greetingForHour(hour: number) {
  if (hour >= 5 && hour < 12) return "Good morning";
  if (hour >= 12 && hour < 17) return "Good afternoon";
  return "Good evening";
}

