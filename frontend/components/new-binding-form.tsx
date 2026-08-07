"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";

const PROVIDERS = [
  { value: "anthropic", label: "Anthropic (Claude)", defaultModel: "claude-sonnet-5" },
  { value: "litellm", label: "LiteLLM (URL + token)", defaultModel: "gpt-4o" },
];

export function NewBindingForm() {
  const router = useRouter();
  const [provider, setProvider] = useState("anthropic");
  const [modelName, setModelName] = useState("claude-sonnet-5");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isLiteLLM = provider === "litellm";

  function handleProviderChange(value: string) {
    setProvider(value);
    setModelName(PROVIDERS.find((p) => p.value === value)?.defaultModel ?? "");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await api.createBinding({
        provider,
        model_name: modelName,
        api_key: apiKey,
        ...(isLiteLLM ? { config: { base_url: baseUrl } } : {}),
      });
      setApiKey("");
      setBaseUrl("");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add binding");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Card className="p-6">
      <h2 className="mb-4 text-sm font-semibold text-foreground">Add a model binding</h2>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-muted">Provider</span>
            <select
              value={provider}
              onChange={(e) => handleProviderChange(e.target.value)}
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
            <Input
              required
              value={modelName}
              onChange={(e) => setModelName(e.target.value)}
              placeholder={isLiteLLM ? "gpt-4o" : "claude-sonnet-5"}
            />
          </label>
        </div>

        {isLiteLLM && (
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-muted">Base URL</span>
            <Input
              required
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder="https://your-litellm-proxy.example.com"
            />
            <span className="text-xs text-muted">
              Your LiteLLM proxy&apos;s base URL — requests go to {"{base URL}"}/chat/completions.
            </span>
          </label>
        )}

        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">{isLiteLLM ? "Token" : "API key"}</span>
          <Input
            required
            type="password"
            autoComplete="off"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder={isLiteLLM ? "sk-..." : "sk-ant-..."}
          />
          <span className="text-xs text-muted">
            Stored encrypted at rest (AES-256-GCM), never shown again after saving.
          </span>
        </label>
        {error && <p className="text-sm text-danger">{error}</p>}
        <div className="flex justify-end">
          <Button type="submit" disabled={submitting || !apiKey || (isLiteLLM && !baseUrl)}>
            {submitting ? "Saving…" : "Save binding"}
          </Button>
        </div>
      </form>
    </Card>
  );
}
