import { cn } from "@/lib/cn";
import type { TextareaHTMLAttributes } from "react";

export function Textarea({
  className,
  ...props
}: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "w-full resize-none rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm text-foreground placeholder:text-muted transition-colors focus:outline-none focus:ring-2 focus:ring-ring",
        className
      )}
      {...props}
    />
  );
}
