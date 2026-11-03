import { NextResponse } from "next/server";

import { prisma } from "@/lib/db";

export async function GET() {
  const items = await prisma.item.findMany({
    orderBy: { createdAt: "desc" },
  });
  return NextResponse.json(items);
}

export async function POST(request: Request) {
  const payload = (await request.json()) as { name?: unknown };
  const name = typeof payload.name === "string" ? payload.name.trim() : "";

  if (!name) {
    return NextResponse.json({ error: "name is required" }, { status: 400 });
  }

  const item = await prisma.item.create({ data: { name } });
  return NextResponse.json(item, { status: 201 });
}
