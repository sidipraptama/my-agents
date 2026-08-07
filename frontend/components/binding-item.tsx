"use client";

import { Pencil, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import type { ModelBinding } from "@/lib/types";

const PROVIDERS = [
  { value: "anthropic", label: "Anthropic (Claude)" },
  { value: "litellm", label: "LiteLLM (URL + token)" },
];

function baseUrlOf(binding: ModelBinding) {
  return typeof binding.config.base_url === "string" ? binding.config.base_url : "";
}

export function BindingItem({ binding }: { binding: ModelBinding }) {
  const router = useRouter();
  const [editing, setEditing] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [provider, setProvider] = useState(binding.provider);
  const [modelName, setModelName] = useState(binding.model_name);
  const [baseUrl, setBaseUrl] = useState(baseUrlOf(binding));
  const [apiKey, setApiKey] = useState("");
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isLiteLLM = provider === "litellm";

  function cancelEdit() {
    setEditing(false);
    setProvider(binding.provider);
    setModelName(binding.model_name);
    setBaseUrl(baseUrlOf(binding));
    setApiKey("");
    setError(null);
  }

  async function handleSave() {
    setSaving(true);
    setError(null);
    try {
      await api.updateBinding(binding.id, {
        provider,
        model_name: modelName,
        config: isLiteLLM ? { base_url: baseUrl } : {},
        ...(apiKey ? { api_key: apiKey } : {}),
      });
      setEditing(false);
      setApiKey("");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update binding");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    setDeleting(true);
    try {
      await api.deleteBinding(binding.id);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete binding");
      setDeleting(false);
    }
  }

  if (!editing) {
    return (
      <Card className="flex items-center justify-between p-4">
        <div>
          <p className="text-sm font-medium text-foreground">{binding.model_name}</p>
          <p className="text-xs text-muted">
            {binding.provider}
            {baseUrlOf(binding) ? ` · ${baseUrlOf(binding)}` : ""}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Badge tone="success">Connected</Badge>
          <Button variant="ghost" onClick={() => setEditing(true)} aria-label="Edit binding">
            <Pencil size={14} />
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <Card className="flex flex-col gap-4 p-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Provider</span>
          <select
            value={provider}
            onChange={(e) => setProvider(e.target.value)}
            className="w-full rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          >
            {PROVIDERS.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Model name</span>
          <Input value={modelName} onChange={(e) => setModelName(e.target.value)} />
        </label>
      </div>

      {isLiteLLM && (
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Base URL</span>
          <Input value={baseUrl} onChange={(e) => setBaseUrl(e.target.value)} />
        </label>
      )}

      <label className="flex flex-col gap-1.5 text-sm">
        <span className="text-muted">
          {isLiteLLM ? "Token" : "API key"} (leave blank to keep the current one)
        </span>
        <Input
          type="password"
          autoComplete="off"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="••••••••"
        />
      </label>

      {error && <p className="text-sm text-danger">{error}</p>}

      <div className="flex items-center justify-between">
        {confirmingDelete ? (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted">Delete this binding?</span>
            <Button variant="secondary" onClick={() => setConfirmingDelete(false)}>
              Cancel
            </Button>
            <Button variant="danger" onClick={handleDelete} disabled={deleting}>
              {deleting ? "Deleting…" : "Confirm"}
            </Button>
          </div>
        ) : (
          <Button variant="ghost" onClick={() => setConfirmingDelete(true)}>
            <Trash2 size={14} />
            Delete
          </Button>
        )}
        <div className="flex gap-2">
          <Button variant="secondary" onClick={cancelEdit}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !modelName || (isLiteLLM && !baseUrl)}>
            {saving ? "Saving…" : "Save"}
          </Button>
        </div>
      </div>
    </Card>
  );
}
