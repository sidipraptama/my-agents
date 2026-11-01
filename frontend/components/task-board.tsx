"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { api } from "@/lib/api";
import { cn } from "@/lib/cn";
import type { TaskStatus, TaskWithAgent } from "@/lib/types";

const STATUS_ORDER: TaskStatus[] = ["running", "waiting_approval", "done", "rejected", "failed"];

const STATUS_TONE: Record<TaskStatus, "neutral" | "warning" | "success" | "danger"> = {
  queued: "neutral",
  running: "warning",
  waiting_approval: "warning",
  done: "success",
  rejected: "neutral",
  failed: "danger",
};

const STATUS_LABEL: Record<TaskStatus, string> = {
  queued: "queued",
  running: "running",
  waiting_approval: "needs approval",
  done: "done",
  rejected: "rejected",
  failed: "failed",
};

function TriggerBadge({ trigger }: { trigger: string }) {
  return <Badge tone="neutral">{trigger}</Badge>;
}

function ApprovalActions({ taskId }: { taskId: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState<"approve" | "reject" | null>(null);

  async function handle(action: "approve" | "reject") {
    setBusy(action);
    try {
      if (action === "approve") await api.approveTask(taskId);
      else await api.rejectTask(taskId);
      router.refresh();
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="flex gap-2" onClick={(e) => e.stopPropagation()}>
      <Button variant="secondary" onClick={() => handle("reject")} disabled={busy !== null}>
        {busy === "reject" ? "Rejecting…" : "Reject"}
      </Button>
      <Button onClick={() => handle("approve")} disabled={busy !== null}>
        {busy === "approve" ? "Approving…" : "Approve"}
      </Button>
    </div>
  );
}

function TaskRow({ task }: { task: TaskWithAgent }) {
  const needsApproval = task.status === "waiting_approval";

  const body = (
    <>
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-foreground">{task.agent_name}</span>
          <TriggerBadge trigger={task.trigger_type} />
        </div>
        <p className="truncate text-xs text-muted">{task.input}</p>
      </div>
      <div className="flex shrink-0 items-center gap-3">
        <span className="text-xs text-muted">{new Date(task.created_at).toLocaleString()}</span>
        <Badge tone={STATUS_TONE[task.status]}>{STATUS_LABEL[task.status]}</Badge>
        {needsApproval && <ApprovalActions taskId={task.id} />}
      </div>
    </>
  );

  if (needsApproval) {
    return (
      <div className="flex items-center justify-between gap-4 rounded-xl border border-border bg-surface p-3">
        {body}
      </div>
    );
  }

  return (
    <Link
      href={`/agents/${task.agent_id}`}
      className="flex items-center justify-between gap-4 rounded-xl border border-border bg-surface p-3 transition-colors hover:bg-surface-hover"
    >
      {body}
    </Link>
  );
}

function TaskCard({ task }: { task: TaskWithAgent }) {
  const needsApproval = task.status === "waiting_approval";

  const body = (
    <>
      <div className="flex items-center justify-between gap-2">
        <span className="truncate text-sm font-medium text-foreground">{task.agent_name}</span>
        <TriggerBadge trigger={task.trigger_type} />
      </div>
      <p className="line-clamp-2 text-xs text-muted">{task.input}</p>
      <span className="text-[11px] text-muted">{new Date(task.created_at).toLocaleString()}</span>
      {needsApproval && <ApprovalActions taskId={task.id} />}
    </>
  );

  if (needsApproval) {
    return (
      <div className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3">{body}</div>
    );
  }

  return (
    <Link
      href={`/agents/${task.agent_id}`}
      className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3 transition-colors hover:bg-surface-hover"
    >
      {body}
    </Link>
  );
}

export function TaskBoard({ tasks }: { tasks: TaskWithAgent[] }) {
  const [view, setView] = useState<"list" | "board">("board");

  if (tasks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-border py-24 text-center">
        <p className="text-sm text-muted">No tasks yet.</p>
        <p className="text-xs text-muted">
          Chat with an Agent, or trigger one via a schedule, webhook, or the API.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end gap-1 rounded-full border border-border bg-surface p-1 self-end">
        {(["board", "list"] as const).map((v) => (
          <button
            key={v}
            onClick={() => setView(v)}
            className={cn(
              "rounded-full px-3 py-1 text-xs font-medium capitalize transition-colors",
              view === v ? "bg-accent/10 text-accent" : "text-muted hover:text-foreground"
            )}
          >
            {v}
          </button>
        ))}
      </div>

      {view === "list" ? (
        <div className="flex flex-col gap-2">
          {tasks.map((t) => (
            <TaskRow key={t.id} task={t} />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
          {STATUS_ORDER.map((status) => {
            const columnTasks = tasks.filter((t) => t.status === status);
            return (
              <div key={status} className="flex flex-col gap-2">
                <div className="flex items-center justify-between px-1">
                  <span className="text-xs font-semibold uppercase tracking-wide text-muted">
                    {STATUS_LABEL[status]}
                  </span>
                  <Badge tone={STATUS_TONE[status]}>{columnTasks.length}</Badge>
                </div>
                <div className="scrollbar-thin flex max-h-[calc(100vh-14rem)] flex-col gap-2 overflow-y-auto">
                  {columnTasks.map((t) => (
                    <TaskCard key={t.id} task={t} />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
