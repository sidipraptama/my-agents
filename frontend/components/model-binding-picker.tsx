"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import type { ModelBinding } from "@/lib/types";

export function ModelBindingPicker({
  agentId,
  currentBindingId,
  bindings,
}: {
  agentId: string;
  currentBindingId: string | null;
  bindings: ModelBinding[];
}) {
  const router = useRouter();
  const [selected, setSelected] = useState(currentBindingId ?? "");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (bindings.length === 0) {
    return (
      <p className="text-sm text-muted">
        No model bindings yet.{" "}
        <Link href="/bindings" className="text-accent hover:underline">
          Add one
        </Link>{" "}
        to attach a model to this agent.
      </p>
    );
  }

  async function handleSave() {
    setSaving(true);
    setError(null);
    try {
      await api.updateAgent(agentId, { model_binding_id: selected || null });
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update model binding");
    } finally {
      setSaving(false);
    }
  }

  const dirty = selected !== (currentBindingId ?? "");

  return (
    <div className="flex flex-col gap-2">
      <select
        value={selected}
        onChange={(e) => setSelected(e.target.value)}
        className="w-full rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
      >
        <option value="">Not configured</option>
        {bindings.map((b) => (
          <option key={b.id} value={b.id}>
            {b.provider} · {b.model_name}
          </option>
        ))}
      </select>
      {error && <p className="text-sm text-danger">{error}</p>}
      {dirty && (
        <Button onClick={handleSave} disabled={saving} className="w-full">
          {saving ? "Saving…" : "Save"}
        </Button>
      )}
    </div>
  );
}
