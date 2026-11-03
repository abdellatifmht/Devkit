import Link from "next/link";

export default function NotFound() {
  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col items-start justify-center gap-4 px-6">
      <p className="text-sm font-semibold uppercase tracking-[0.2em] text-sky-600">404</p>
      <h1 className="text-3xl font-bold text-slate-900">Page not found</h1>
      <Link className="text-sky-700 hover:underline" href="/">
        Back to the home page
      </Link>
    </main>
  );
}
