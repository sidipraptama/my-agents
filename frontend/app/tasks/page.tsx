import { TaskBoard } from "@/components/task-board";
import { api } from "@/lib/api";

export default async function TasksPage() {
  const tasks = await api.listAllTasks().catch(() => []);

  return (
    <div className="mx-auto max-w-6xl px-8 py-10">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight">Task Board</h1>
        <p className="mt-1 text-sm text-muted">
          Every task across every Agent — manual, scheduled, webhook, and API-triggered.
        </p>
      </div>

      <TaskBoard tasks={tasks} />
    </div>
  );
}
