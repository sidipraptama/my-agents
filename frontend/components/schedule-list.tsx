"use client";

import { Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import type { Schedule } from "@/lib/types";

export function ScheduleList({ agentId, schedules }: { agentId: string; schedules: Schedule[] }) {
  const router = useRouter();
  const [adding, setAdding] = useState(false);
  const [cronExpr, setCronExpr] = useState("0 9 * * *");
  const [input, setInput] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleCreate() {
    setSubmitting(true);
    setError(null);
    try {
      await api.createSchedule(agentId, { cron_expr: cronExpr, input });
      setInput("");
      setAdding(false);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create schedule");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleToggle(schedule: Schedule) {
    await api.updateSchedule(schedule.id, {
      cron_expr: schedule.cron_expr,
      input: schedule.input,
      enabled: !schedule.enabled,
    });
    router.refresh();
  }

  async function handleDelete(id: string) {
    await api.deleteSchedule(id);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-3">
      {schedules.map((s) => (
        <div
          key={s.id}
          className="flex flex-col gap-1.5 rounded-xl border border-border bg-surface p-3"
        >
          <div className="flex items-center justify-between gap-2">
            <code className="truncate text-xs text-foreground">{s.cron_expr}</code>
            <div className="flex shrink-0 items-center gap-2">
              <Badge tone={s.enabled ? "success" : "neutral"}>{s.enabled ? "On" : "Off"}</Badge>
              <button
                onClick={() => handleToggle(s)}
                className="text-xs text-accent hover:underline"
              >
                {s.enabled ? "Pause" : "Resume"}
              </button>
              <button
                onClick={() => handleDelete(s.id)}
                aria-label="Delete schedule"
                className="text-muted hover:text-danger"
              >
                <Trash2 size={13} />
              </button>
            </div>
          </div>
          <p className="line-clamp-2 text-xs text-muted">{s.input}</p>
        </div>
      ))}

      {adding ? (
        <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3">
          <label className="flex flex-col gap-1 text-xs">
            <span className="text-muted">Cron expression</span>
            <Input
              value={cronExpr}
              onChange={(e) => setCronExpr(e.target.value)}
              placeholder="0 9 * * *"
            />
          </label>
          <label className="flex flex-col gap-1 text-xs">
            <span className="text-muted">Message to send</span>
            <Textarea
              rows={2}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Summarize overnight alerts"
            />
          </label>
          {error && <p className="text-xs text-danger">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={() => setAdding(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} disabled={submitting || !cronExpr || !input}>
              {submitting ? "Adding…" : "Add"}
            </Button>
          </div>
        </div>
      ) : (
        <Button variant="secondary" onClick={() => setAdding(true)}>
          + Add schedule
        </Button>
      )}
    </div>
  );
}
