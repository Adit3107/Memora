import { NextResponse } from "next/server";

export async function POST(req: Request) {
  try {
    const payload = await req.json();
    const eventType = payload?.type || "unknown";

    // Log received Clerk webhook event
    console.log(`[CLERK WEBHOOK] Received event: ${eventType}`, {
      id: payload?.data?.id,
      email: payload?.data?.email_addresses?.[0]?.email_address,
      status: payload?.data?.verification?.status,
    });

    // In production or full sync, verify signature with svix using process.env.CLERK_WEBHOOK_SECRET
    // and sync user profile / email verification state with backend PostgreSQL
    if (eventType === "user.created" || eventType === "user.updated") {
      const email = payload?.data?.email_addresses?.[0]?.email_address;
      const isEmailVerified = payload?.data?.email_addresses?.[0]?.verification?.status === "verified";
      console.log(`[CLERK WEBHOOK] User synced: ${email}, verified: ${isEmailVerified}`);
    }

    return NextResponse.json({ success: true, event: eventType });
  } catch (error) {
    console.error("[CLERK WEBHOOK ERROR]", error);
    return NextResponse.json(
      { success: false, error: "Failed to process webhook" },
      { status: 400 }
    );
  }
}
