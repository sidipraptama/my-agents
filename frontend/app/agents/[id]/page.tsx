import { notFound } from "next/navigation";

import { ChatWindow } from "@/components/chat-window";
import { DeleteAgentButton } from "@/components/delete-agent-button";
import { ModelBindingPicker } from "@/components/model-binding-picker";
import { ScheduleList } from "@/components/schedule-list";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { WebhookList } from "@/components/webhook-list";
import { api } from "@/lib/api";

const PUBLIC_API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export default async function AgentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  const [agent, tasks, bindings, schedules, webhooks] = await Promise.all([
    api.getAgent(id).catch(() => null),
    api.listTasks(id).catch(() => []),
    api.listBindings().catch(() => []),
    api.listSchedules(id).catch(() => []),
    api.listWebhooks(id).catch(() => []),
  ]);

  if (!agent) notFound();

  return (
    <div className="flex h-full">
      <div className="scrollbar-thin w-80 shrink-0 overflow-y-auto border-r border-border p-6">
        <div className="mb-4 flex items-center justify-between">
          <h1 className="text-lg font-semibold tracking-tight">{agent.name}</h1>
          <Badge tone={agent.enabled ? "success" : "neutral"}>
            {agent.enabled ? "Online" : "Paused"}
          </Badge>
        </div>
        <p className="mb-6 text-sm text-muted">{agent.role || "No role set"}</p>

        <Card className="mb-4 flex flex-col gap-3 p-4">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-muted">
            Persona
          </h2>
          {agent.persona.bio && <p className="text-sm text-foreground">{agent.persona.bio}</p>}
          {agent.persona.tone && (
            <p className="text-sm text-muted">Tone: {agent.persona.tone}</p>
          )}
          {agent.persona.knowledge.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-muted">Knowledge</p>
              <ul className="list-inside list-disc text-sm text-foreground">
                {agent.persona.knowledge.map((k, i) => (
                  <li key={i}>{k}</li>
                ))}
              </ul>
            </div>
          )}
          {agent.persona.constraints.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-muted">Constraints</p>
              <ul className="list-inside list-disc text-sm text-foreground">
                {agent.persona.constraints.map((c, i) => (
                  <li key={i}>{c}</li>
                ))}
              </ul>
            </div>
          )}
        </Card>

        <Card className="mb-4 p-4">
          <h2 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted">
            Model
          </h2>
          <ModelBindingPicker
            agentId={agent.id}
            currentBindingId={agent.model_binding_id}
            bindings={bindings}
          />
        </Card>

        <Card className="mb-4 p-4">
          <h2 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">
            Schedules
          </h2>
          <ScheduleList agentId={agent.id} schedules={schedules} />
        </Card>

        <Card className="mb-4 p-4">
          <h2 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted">
            Webhooks
          </h2>
          <WebhookList agentId={agent.id} webhooks={webhooks} />
        </Card>

        <Card className="mb-4 p-4">
          <h2 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted">
            API
          </h2>
          <p className="mb-2 text-xs text-muted">Trigger a task programmatically:</p>
          <pre className="scrollbar-thin overflow-x-auto rounded-lg bg-surface p-2 text-xs text-foreground">
            {`curl -X POST ${PUBLIC_API_URL}/api/agents/${agent.id}/tasks \\\n  -H "Content-Type: application/json" \\\n  -d '{"message": "..."}'`}
          </pre>
        </Card>

        <div className="mt-6">
          <DeleteAgentButton agentId={agent.id} />
        </div>
      </div>

      <div className="flex-1">
        <ChatWindow
          agentId={agent.id}
          initialTasks={tasks}
          hasModel={!!agent.model_binding_id}
        />
      </div>
    </div>
  );
}
