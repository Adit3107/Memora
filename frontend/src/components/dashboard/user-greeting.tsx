"use client";

import { useUser } from "@clerk/nextjs";
import { useEffect, useState } from "react";

export function UserGreeting() {
  const { user, isLoaded } = useUser();
  const [greeting, setGreeting] = useState("Good day");

  useEffect(() => {
    const hour = new Date().getHours();
    if (hour < 12) setGreeting("Good morning");
    else if (hour < 17) setGreeting("Good afternoon");
    else setGreeting("Good evening");
  }, []);

  const name =
    user?.firstName ||
    (user?.fullName ? user.fullName.split(" ")[0] : null) ||
    "there";

  return (
    <h1 className="text-3xl font-semibold tracking-normal text-foreground sm:text-4xl">
      {isLoaded && user ? `${greeting}, ${name}.` : `${greeting}.`}
    </h1>
  );
}
