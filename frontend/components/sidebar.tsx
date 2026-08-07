"use client";

import { Bot, KeyRound, ListChecks } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/cn";

const NAV = [
  { href: "/agents", label: "Agents", icon: Bot },
  { href: "/tasks", label: "Task Board", icon: ListChecks },
  { href: "/bindings", label: "Model Bindings", icon: KeyRound },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="flex h-full w-60 shrink-0 flex-col border-r border-border bg-surface/70 backdrop-blur-xl">
      <div className="flex items-center gap-2 px-5 py-5">
        <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-accent text-sm font-semibold text-accent-foreground">
          m
        </div>
        <span className="text-[15px] font-semibold tracking-tight">my-agents</span>
      </div>

      <nav className="flex flex-1 flex-col gap-0.5 px-3">
        {NAV.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                active
                  ? "bg-accent/10 text-accent"
                  : "text-muted hover:bg-surface-hover hover:text-foreground"
              )}
            >
              <Icon size={16} strokeWidth={2} />
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="flex items-center justify-between border-t border-border px-5 py-4">
        <span className="text-xs text-muted">Self-hosted</span>
        <ThemeToggle />
      </div>
    </aside>
  );
}
