"use client";

import { Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import type { RedLineRule } from "@/lib/types";

export function RedLineRules({ agentId, rules }: { agentId: string; rules: RedLineRule[] }) {
  const router = useRouter();
  const [adding, setAdding] = useState(false);
  const [pattern, setPattern] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleCreate() {
    setSubmitting(true);
    try {
      await api.createRedLineRule(agentId, pattern);
      setPattern("");
      setAdding(false);
      router.refresh();
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    await api.deleteRedLineRule(id);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-3">
      <p className="text-xs text-muted">
        A task pauses for your approval whenever its message contains one of these (case-insensitive).
      </p>

      {rules.length === 0 && !adding && <p className="text-xs text-muted">No rules yet.</p>}

      {rules.map((rule) => (
        <div
          key={rule.id}
          className="flex items-center justify-between gap-2 rounded-xl border border-border bg-surface p-3"
        >
          <code className="truncate text-xs text-foreground">{rule.pattern}</code>
          <button
            onClick={() => handleDelete(rule.id)}
            aria-label="Delete rule"
            className="shrink-0 text-muted hover:text-danger"
          >
            <Trash2 size={13} />
          </button>
        </div>
      ))}

      {adding ? (
        <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3">
          <Input
            value={pattern}
            onChange={(e) => setPattern(e.target.value)}
            placeholder="delete"
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setAdding(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={submitting || !pattern.trim()}>
              {submitting ? "Adding…" : "Add"}
            </Button>
          </div>
        </div>
      ) : (
        <Button variant="secondary" onClick={() => setAdding(true)}>
          + Add rule
        </Button>
      )}
    </div>
  );
}
