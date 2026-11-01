"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { api } from "@/lib/api";

export function AutoRememberToggle({
  agentId,
  autoRemember,
}: {
  agentId: string;
  autoRemember: boolean;
}) {
  const router = useRouter();
  const [saving, setSaving] = useState(false);

  async function handleToggle() {
    setSaving(true);
    try {
      await api.updateAgent(agentId, { auto_remember: !autoRemember });
      router.refresh();
    } finally {
      setSaving(false);
    }
  }

  return (
    <label className="flex cursor-pointer items-start gap-2.5 text-sm">
      <input
        type="checkbox"
        checked={autoRemember}
        onChange={handleToggle}
        disabled={saving}
        className="mt-0.5 h-4 w-4 accent-accent"
      />
      <span className="flex flex-col">
        <span className="font-medium text-foreground">Auto-remember</span>
        <span className="text-xs text-muted">
          After each task, a follow-up model call checks whether anything is worth remembering
          long-term. Off by default — this roughly doubles model calls per task.
        </span>
      </span>
    </label>
  );
}
