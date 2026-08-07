import Link from "next/link";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { AgentCard } from "@/components/agent-card";
import { api } from "@/lib/api";

export default async function AgentsPage() {
  const agents = await api.listAgents().catch(() => []);

  return (
    <div className="mx-auto max-w-5xl px-8 py-10">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Agents</h1>
          <p className="mt-1 text-sm text-muted">
            Persistent agents with a persona, a role, and a model of their own.
          </p>
        </div>
        <Link href="/agents/new">
          <Button>
            <Plus size={16} />
            New Agent
          </Button>
        </Link>
      </div>

      {agents.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-border py-24 text-center">
          <p className="text-sm text-muted">No agents yet.</p>
          <Link href="/agents/new">
            <Button variant="secondary">Create your first Agent</Button>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {agents.map((agent) => (
            <AgentCard key={agent.id} agent={agent} />
          ))}
        </div>
      )}
    </div>
  );
}
