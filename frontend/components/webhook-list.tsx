"use client";

import { Check, Copy, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import type { Webhook } from "@/lib/types";

const PUBLIC_API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function webhookUrl(token: string) {
  return `${PUBLIC_API_URL}/api/hooks/${token}`;
}

export function WebhookList({ agentId, webhooks }: { agentId: string; webhooks: Webhook[] }) {
  const router = useRouter();
  const [creating, setCreating] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  async function handleCreate() {
    setCreating(true);
    try {
      await api.createWebhook(agentId);
      router.refresh();
    } finally {
      setCreating(false);
    }
  }

  async function handleCopy(webhook: Webhook) {
    await navigator.clipboard.writeText(webhookUrl(webhook.token));
    setCopiedId(webhook.id);
    setTimeout(() => setCopiedId((current) => (current === webhook.id ? null : current)), 1500);
  }

  async function handleDelete(id: string) {
    await api.deleteWebhook(id);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-3">
      {webhooks.map((wh) => (
        <div
          key={wh.id}
          className="flex items-center justify-between gap-2 rounded-xl border border-border bg-surface p-3"
        >
          <code className="truncate text-xs text-foreground">{webhookUrl(wh.token)}</code>
          <div className="flex shrink-0 items-center gap-2">
            <button
              onClick={() => handleCopy(wh)}
              aria-label="Copy webhook URL"
              className="text-muted hover:text-foreground"
            >
              {copiedId === wh.id ? <Check size={13} /> : <Copy size={13} />}
            </button>
            <button
              onClick={() => handleDelete(wh.id)}
              aria-label="Delete webhook"
              className="text-muted hover:text-danger"
            >
              <Trash2 size={13} />
            </button>
          </div>
        </div>
      ))}

      <Button variant="secondary" onClick={handleCreate} disabled={creating}>
        {creating ? "Creating…" : "+ Add webhook"}
      </Button>
      <p className="text-xs text-muted">
        POST JSON like <code>{`{"message": "..."}`}</code> (or plain text) to trigger a task.
      </p>
    </div>
  );
}
