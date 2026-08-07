import { NewAgentForm } from "@/components/new-agent-form";
import { api } from "@/lib/api";

export default async function NewAgentPage() {
  const bindings = await api.listBindings().catch(() => []);

  return (
    <div className="mx-auto max-w-2xl px-8 py-10">
      <h1 className="mb-8 text-2xl font-semibold tracking-tight">New Agent</h1>
      <NewAgentForm bindings={bindings} />
    </div>
  );
}
