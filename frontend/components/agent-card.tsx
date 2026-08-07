import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import type { Agent } from "@/lib/types";

function initial(name: string) {
  return name.trim().charAt(0).toUpperCase() || "?";
}

export function AgentCard({ agent }: { agent: Agent }) {
  return (
    <Link href={`/agents/${agent.id}`}>
      <Card className="flex h-full flex-col gap-3 p-5 transition-transform hover:-translate-y-0.5 hover:shadow-md">
        <div className="flex items-start justify-between">
          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-accent/10 text-base font-semibold text-accent">
            {initial(agent.name)}
          </div>
          <Badge tone={agent.enabled ? "success" : "neutral"}>
            {agent.enabled ? "Online" : "Paused"}
          </Badge>
        </div>
        <div>
          <h3 className="text-[15px] font-semibold text-foreground">{agent.name}</h3>
          <p className="text-sm text-muted">{agent.role || "No role set"}</p>
        </div>
        {agent.persona.bio && (
          <p className="line-clamp-2 text-sm text-muted">{agent.persona.bio}</p>
        )}
        <div className="mt-auto pt-1 text-xs text-muted">
          {agent.model_binding_id ? "Model connected" : "No model bound"}
        </div>
      </Card>
    </Link>
  );
}
