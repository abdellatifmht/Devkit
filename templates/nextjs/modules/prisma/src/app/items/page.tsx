import Link from "next/link";

import { prisma } from "@/lib/db";

export const dynamic = "force-dynamic";

export default async function ItemsPage() {
  const items = await prisma.item.findMany({
    orderBy: { createdAt: "desc" },
  });

  return (
    <main className="mx-auto max-w-3xl px-6 py-16">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.2em] text-sky-600">
            Prisma
          </p>
          <h1 className="mt-2 text-3xl font-bold text-slate-900">Items</h1>
        </div>
        <Link className="text-sm font-medium text-sky-700 hover:underline" href="/">
          Back home
        </Link>
      </div>

      {items.length === 0 ? (
        <p className="mt-8 rounded-xl border border-dashed border-slate-300 bg-white p-6 text-slate-600">
          No items yet. Send a POST request to{" "}
          <code className="rounded bg-slate-100 px-1.5 py-0.5">/api/items</code> to
          create one.
        </p>
      ) : (
        <ul className="mt-8 divide-y divide-slate-200 rounded-xl border border-slate-200 bg-white">
          {items.map((item) => (
            <li className="flex items-center justify-between gap-4 px-5 py-4" key={item.id}>
              <span className="font-medium text-slate-900">{item.name}</span>
              <time className="text-sm text-slate-500" dateTime={item.createdAt.toISOString()}>
                {item.createdAt.toLocaleDateString("en-US")}
              </time>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
