"use client";

import { Pencil, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import type { MemoryEntry, MemoryEntryType } from "@/lib/types";

const TYPES: { value: MemoryEntryType; label: string; tone: "neutral" | "warning" | "success" }[] = [
  { value: "fact", label: "Fact", tone: "neutral" },
  { value: "preference", label: "Preference", tone: "warning" },
  { value: "summary", label: "Auto", tone: "success" },
];

function toneFor(type: MemoryEntryType) {
  return TYPES.find((t) => t.value === type)?.tone ?? "neutral";
}

function labelFor(type: MemoryEntryType) {
  return TYPES.find((t) => t.value === type)?.label ?? type;
}

function MemoryRow({ entry }: { entry: MemoryEntry }) {
  const router = useRouter();
  const [editing, setEditing] = useState(false);
  const [entryType, setEntryType] = useState<MemoryEntryType>(entry.entry_type);
  const [content, setContent] = useState(entry.content);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  async function handleSave() {
    setSaving(true);
    try {
      await api.updateMemory(entry.id, { entry_type: entryType, content });
      setEditing(false);
      router.refresh();
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    setDeleting(true);
    try {
      await api.deleteMemory(entry.id);
      router.refresh();
    } finally {
      setDeleting(false);
    }
  }

  if (editing) {
    return (
      <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3">
        <select
          value={entryType}
          onChange={(e) => setEntryType(e.target.value as MemoryEntryType)}
          className="w-fit rounded-lg border border-border bg-surface px-2 py-1 text-xs text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        >
          {TYPES.filter((t) => t.value !== "summary").map((t) => (
            <option key={t.value} value={t.value}>
              {t.label}
            </option>
          ))}
        </select>
        <Textarea rows={2} value={content} onChange={(e) => setContent(e.target.value)} />
        <div className="flex justify-end gap-2">
          <Button variant="secondary" onClick={() => setEditing(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !content.trim()}>
            {saving ? "Saving…" : "Save"}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-start justify-between gap-2 rounded-xl border border-border bg-surface p-3">
      <div className="flex min-w-0 flex-col gap-1">
        <Badge tone={toneFor(entry.entry_type)}>{labelFor(entry.entry_type)}</Badge>
        <p className="text-xs text-foreground">{entry.content}</p>
      </div>
      <div className="flex shrink-0 items-center gap-2">
        <button
          onClick={() => setEditing(true)}
          aria-label="Edit memory"
          className="text-muted hover:text-foreground"
        >
          <Pencil size={13} />
        </button>
        <button
          onClick={handleDelete}
          disabled={deleting}
          aria-label="Delete memory"
          className="text-muted hover:text-danger"
        >
          <Trash2 size={13} />
        </button>
      </div>
    </div>
  );
}

export function MemoryList({ agentId, entries }: { agentId: string; entries: MemoryEntry[] }) {
  const router = useRouter();
  const [adding, setAdding] = useState(false);
  const [entryType, setEntryType] = useState<MemoryEntryType>("fact");
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleCreate() {
    setSubmitting(true);
    try {
      await api.createMemory(agentId, { entry_type: entryType, content });
      setContent("");
      setAdding(false);
      router.refresh();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex flex-col gap-3">
      {entries.length === 0 && !adding && (
        <p className="text-xs text-muted">Nothing remembered yet.</p>
      )}

      {entries.map((entry) => (
        <MemoryRow key={entry.id} entry={entry} />
      ))}

      {adding ? (
        <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3">
          <select
            value={entryType}
            onChange={(e) => setEntryType(e.target.value as MemoryEntryType)}
            className="w-fit rounded-lg border border-border bg-surface px-2 py-1 text-xs text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          >
            {TYPES.filter((t) => t.value !== "summary").map((t) => (
              <option key={t.value} value={t.value}>
                {t.label}
              </option>
            ))}
          </select>
          <Textarea
            rows={2}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="Prefers concise answers, no filler."
          />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setAdding(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={submitting || !content.trim()}>
              {submitting ? "Adding…" : "Add"}
            </Button>
          </div>
        </div>
      ) : (
        <Button variant="secondary" onClick={() => setAdding(true)}>
          + Add memory
        </Button>
      )}
    </div>
  );
}
