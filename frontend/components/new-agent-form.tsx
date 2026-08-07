"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { cn } from "@/lib/cn";
import { AGENT_TEMPLATES, type AgentTemplate } from "@/lib/templates";
import type { ModelBinding } from "@/lib/types";

function listFromLines(value: string): string[] {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean);
}

export function NewAgentForm({ bindings }: { bindings: ModelBinding[] }) {
  const router = useRouter();
  const [templateId, setTemplateId] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [role, setRole] = useState("");
  const [bio, setBio] = useState("");
  const [tone, setTone] = useState("");
  const [constraints, setConstraints] = useState("");
  const [knowledge, setKnowledge] = useState("");
  const [modelBindingId, setModelBindingId] = useState(bindings[0]?.id ?? "");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function applyTemplate(template: AgentTemplate) {
    setTemplateId(template.id);
    setName(template.suggestedName);
    setRole(template.role);
    setBio(template.persona.bio);
    setTone(template.persona.tone);
    setKnowledge(template.persona.knowledge.join("\n"));
    setConstraints(template.persona.constraints.join("\n"));
  }

  function startBlank() {
    setTemplateId(null);
    setName("");
    setRole("");
    setBio("");
    setTone("");
    setKnowledge("");
    setConstraints("");
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const agent = await api.createAgent({
        name,
        role,
        model_binding_id: modelBindingId || null,
        persona: {
          bio,
          tone,
          constraints: listFromLines(constraints),
          knowledge: listFromLines(knowledge),
        },
      });
      router.push(`/agents/${agent.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create agent");
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <Card className="flex flex-col gap-4 p-6">
        <h2 className="text-sm font-semibold text-foreground">Start from a template</h2>
        <p className="-mt-2 text-sm text-muted">
          Pre-fills role and persona below — everything stays editable after picking one.
        </p>
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <button
            type="button"
            onClick={startBlank}
            className={cn(
              "rounded-xl border px-3 py-3 text-left text-sm transition-colors",
              templateId === null
                ? "border-accent bg-accent/10 text-accent"
                : "border-border bg-surface text-muted hover:bg-surface-hover hover:text-foreground"
            )}
          >
            <span className="block font-medium">Blank</span>
            <span className="block text-xs opacity-80">Start from scratch</span>
          </button>
          {AGENT_TEMPLATES.map((template) => (
            <button
              key={template.id}
              type="button"
              onClick={() => applyTemplate(template)}
              className={cn(
                "rounded-xl border px-3 py-3 text-left text-sm transition-colors",
                templateId === template.id
                  ? "border-accent bg-accent/10 text-accent"
                  : "border-border bg-surface text-foreground hover:bg-surface-hover"
              )}
            >
              <span className="block font-medium">{template.label}</span>
              <span className="block text-xs opacity-70">{template.suggestedName}</span>
            </button>
          ))}
        </div>
      </Card>

      <Card className="flex flex-col gap-4 p-6">
        <h2 className="text-sm font-semibold text-foreground">Identity</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-muted">Name</span>
            <Input
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ada"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm">
            <span className="text-muted">Role</span>
            <Input
              value={role}
              onChange={(e) => setRole(e.target.value)}
              placeholder="Backend Engineer"
            />
          </label>
        </div>
      </Card>

      <Card className="flex flex-col gap-4 p-6">
        <h2 className="text-sm font-semibold text-foreground">Persona</h2>
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Bio</span>
          <Textarea
            rows={2}
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            placeholder="A calm, detail-oriented engineer who double-checks before shipping."
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Tone</span>
          <Input
            value={tone}
            onChange={(e) => setTone(e.target.value)}
            placeholder="Direct, concise, no filler"
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Knowledge (one per line)</span>
          <Textarea
            rows={3}
            value={knowledge}
            onChange={(e) => setKnowledge(e.target.value)}
            placeholder={"Our stack is Go + Next.js + Postgres\nWe deploy via Docker Compose"}
          />
        </label>
        <label className="flex flex-col gap-1.5 text-sm">
          <span className="text-muted">Constraints (one per line)</span>
          <Textarea
            rows={3}
            value={constraints}
            onChange={(e) => setConstraints(e.target.value)}
            placeholder={"Never push directly to main\nAlways ask before deleting data"}
          />
        </label>
      </Card>

      <Card className="flex flex-col gap-4 p-6">
        <h2 className="text-sm font-semibold text-foreground">Model</h2>
        {bindings.length === 0 ? (
          <p className="text-sm text-muted">
            No model bindings yet.{" "}
            <Link href="/bindings" className="text-accent hover:underline">
              Add one
            </Link>{" "}
            to bring your own model (e.g. Claude API).
          </p>
        ) : (
          <select
            value={modelBindingId}
            onChange={(e) => setModelBindingId(e.target.value)}
            className="w-full rounded-xl border border-border bg-surface px-3.5 py-2.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          >
            {bindings.map((b) => (
              <option key={b.id} value={b.id}>
                {b.provider} · {b.model_name}
              </option>
            ))}
          </select>
        )}
      </Card>

      {error && <p className="text-sm text-danger">{error}</p>}

      <div className="flex justify-end gap-3">
        <Link href="/agents">
          <Button type="button" variant="secondary">
            Cancel
          </Button>
        </Link>
        <Button type="submit" disabled={submitting || !name}>
          {submitting ? "Creating…" : "Create Agent"}
        </Button>
      </div>
    </form>
  );
}
