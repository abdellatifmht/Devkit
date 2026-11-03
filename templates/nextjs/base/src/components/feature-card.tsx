import type { ReactNode } from "react";

import { cn } from "@/lib/cn";

type FeatureCardProps = {
  title: string;
  children: ReactNode;
  className?: string;
};

export function FeatureCard({ title, children, className }: FeatureCardProps) {
  return (
    <article className={cn("rounded-xl border border-slate-200 bg-white p-5 shadow-sm", className)}>
      <h2 className="font-semibold text-slate-900">{title}</h2>
      <div className="mt-2 text-sm leading-6 text-slate-600">{children}</div>
    </article>
  );
}
