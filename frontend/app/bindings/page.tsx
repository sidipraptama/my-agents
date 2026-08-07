import { BindingItem } from "@/components/binding-item";
import { Card } from "@/components/ui/card";
import { NewBindingForm } from "@/components/new-binding-form";
import { api } from "@/lib/api";

export default async function BindingsPage() {
  const bindings = await api.listBindings().catch(() => []);

  return (
    <div className="mx-auto max-w-3xl px-8 py-10">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight">Model Bindings</h1>
        <p className="mt-1 text-sm text-muted">
          Bring your own model. Add a provider and API key, then attach it to any Agent.
        </p>
      </div>

      <div className="mb-8 flex flex-col gap-3">
        {bindings.length === 0 ? (
          <Card className="p-6 text-sm text-muted">No model bindings configured yet.</Card>
        ) : (
          bindings.map((b) => <BindingItem key={b.id} binding={b} />)
        )}
      </div>

      <NewBindingForm />
    </div>
  );
}
