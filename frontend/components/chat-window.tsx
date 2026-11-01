"use client";

import { Send } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Markdown } from "@/components/markdown";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { cn } from "@/lib/cn";
import type { Task, TranscriptMessage } from "@/lib/types";

function flatten(tasks: Task[]): TranscriptMessage[] {
  return tasks.flatMap((t) => t.transcript);
}

type Pending = { content: string; status: "sending" | "failed" } | null;

function ThinkingBubble() {
  return (
    <div className="flex justify-start">
      <div className="flex items-center gap-1 rounded-2xl rounded-bl-sm bg-surface-hover px-4 py-3">
        <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted [animation-delay:-0.3s]" />
        <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted [animation-delay:-0.15s]" />
        <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted" />
      </div>
    </div>
  );
}

export function ChatWindow({
  agentId,
  initialTasks,
  hasModel,
}: {
  agentId: string;
  initialTasks: Task[];
  hasModel: boolean;
}) {
  const [tasks, setTasks] = useState(initialTasks);
  const [message, setMessage] = useState("");
  const [pending, setPending] = useState<Pending>(null);
  const [resolving, setResolving] = useState<"approve" | "reject" | null>(null);
  const [error, setError] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  const messages = flatten(tasks);
  const isSending = pending?.status === "sending";
  const lastTask = tasks[tasks.length - 1];
  const waitingApproval = lastTask?.status === "waiting_approval" ? lastTask : null;

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length, pending]);

  async function submit(content: string) {
    setPending({ content, status: "sending" });
    setError(null);
    try {
      const { task } = await api.sendMessage(agentId, content);
      setTasks((prev) => [...prev, task]);
      setPending(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to send message");
      setPending({ content, status: "failed" });
    }
  }

  function handleSend() {
    const trimmed = message.trim();
    if (!trimmed || isSending || waitingApproval) return;
    setMessage("");
    submit(trimmed);
  }

  function handleRetry() {
    if (!pending) return;
    submit(pending.content);
  }

  async function resolveApproval(action: "approve" | "reject") {
    if (!waitingApproval) return;
    setResolving(action);
    setError(null);
    try {
      const { task } =
        action === "approve"
          ? await api.approveTask(waitingApproval.id)
          : await api.rejectTask(waitingApproval.id);
      setTasks((prev) => prev.map((t) => (t.id === task.id ? task : t)));
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${action} task`);
    } finally {
      setResolving(null);
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="scrollbar-thin flex-1 space-y-4 overflow-y-auto px-6 py-6">
        {messages.length === 0 && !pending && (
          <p className="pt-12 text-center text-sm text-muted">
            No messages yet — say hello.
          </p>
        )}
        {messages.map((m, i) => (
          <div
            key={i}
            className={cn("flex", m.role === "user" ? "justify-end" : "justify-start")}
          >
            <div
              className={cn(
                "max-w-[75%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed",
                m.role === "user"
                  ? "whitespace-pre-wrap rounded-br-sm bg-accent text-accent-foreground"
                  : "rounded-bl-sm bg-surface-hover text-foreground"
              )}
            >
              {m.role === "assistant" ? <Markdown content={m.content} /> : m.content}
            </div>
          </div>
        ))}

        {pending && (
          <div className="flex flex-col items-end gap-1">
            <div
              className={cn(
                "max-w-[75%] whitespace-pre-wrap rounded-2xl rounded-br-sm px-4 py-2.5 text-sm leading-relaxed",
                pending.status === "failed"
                  ? "bg-danger/10 text-danger"
                  : "bg-accent text-accent-foreground"
              )}
            >
              {pending.content}
            </div>
            {pending.status === "failed" && (
              <button
                onClick={handleRetry}
                className="pr-1 text-[11px] font-medium text-danger hover:underline"
              >
                Failed to send · Retry
              </button>
            )}
          </div>
        )}

        {waitingApproval && (
          <div className="flex justify-start">
            <div className="flex flex-col gap-2 rounded-2xl rounded-bl-sm bg-surface-hover px-4 py-3">
              <p className="text-xs font-medium text-foreground">
                This message matched a red-line rule — approve it to run, or reject it.
              </p>
              <div className="flex gap-2">
                <Button
                  variant="secondary"
                  onClick={() => resolveApproval("reject")}
                  disabled={resolving !== null}
                >
                  {resolving === "reject" ? "Rejecting…" : "Reject"}
                </Button>
                <Button onClick={() => resolveApproval("approve")} disabled={resolving !== null}>
                  {resolving === "approve" ? "Approving…" : "Approve"}
                </Button>
              </div>
            </div>
          </div>
        )}

        {isSending && <ThinkingBubble />}
        <div ref={bottomRef} />
      </div>

      <div className="border-t border-border p-4">
        {error && <p className="mb-2 text-sm text-danger">{error}</p>}
        {!hasModel && (
          <p className="mb-2 text-sm text-muted">
            This agent has no model bound yet — add one in Model Bindings and attach it in
            Settings before chatting.
          </p>
        )}
        {waitingApproval && (
          <p className="mb-2 text-sm text-muted">Resolve the pending approval above to continue.</p>
        )}
        <div className="flex items-end gap-3">
          <Textarea
            rows={1}
            value={message}
            disabled={!hasModel || !!waitingApproval}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            placeholder="Message this agent…"
            className="max-h-40"
          />
          <Button
            onClick={handleSend}
            disabled={!hasModel || isSending || !message.trim() || !!waitingApproval}
            aria-label="Send"
          >
            <Send size={16} />
          </Button>
        </div>
      </div>
    </div>
  );
}
