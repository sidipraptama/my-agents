import { cn } from "@/lib/cn";
import type { ButtonHTMLAttributes } from "react";

type Variant = "primary" | "secondary" | "ghost" | "danger";

const variants: Record<Variant, string> = {
  primary:
    "bg-accent text-accent-foreground hover:opacity-90 active:opacity-80 shadow-sm",
  secondary:
    "bg-surface text-foreground border border-border hover:bg-surface-hover",
  ghost: "text-muted hover:text-foreground hover:bg-surface-hover",
  danger: "bg-danger text-white hover:opacity-90",
};

export function Button({
  variant = "primary",
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant }) {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-1.5 rounded-full px-4 py-2 text-sm font-medium transition-all duration-150 disabled:opacity-40 disabled:pointer-events-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        variants[variant],
        className
      )}
      {...props}
    />
  );
}
