import { cn } from "@/lib/cn";
import type { InputHTMLAttributes } from "react";

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "w-full rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm text-foreground placeholder:text-muted transition-colors focus:outline-none focus:ring-2 focus:ring-ring",
        className
      )}
      {...props}
    />
  );
}
