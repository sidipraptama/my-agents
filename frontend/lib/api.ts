import type {
  Agent,
  MemoryEntry,
  MemoryEntryType,
  ModelBinding,
  RedLineRule,
  Schedule,
  Task,
  TaskWithAgent,
  Webhook,
} from "./types";

// Server Components run inside the frontend container and must reach the
// backend via the Docker network (INTERNAL_API_URL, read at runtime). The
// browser runs on the host and needs the publicly reachable NEXT_PUBLIC_API_URL,
// which gets inlined into the client bundle at build time.
const API_URL =
  typeof window === "undefined"
    ? (process.env.INTERNAL_API_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080")
    : (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080");

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `Request to ${path} failed with ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  listAgents: () => request<Agent[]>("/api/agents"),
  getAgent: (id: string) => request<Agent>(`/api/agents/${id}`),
  createAgent: (input: Partial<Agent>) =>
    request<Agent>("/api/agents", { method: "POST", body: JSON.stringify(input) }),
  updateAgent: (id: string, input: Partial<Agent>) =>
    request<Agent>(`/api/agents/${id}`, { method: "PATCH", body: JSON.stringify(input) }),
  deleteAgent: (id: string) => request<void>(`/api/agents/${id}`, { method: "DELETE" }),

  listTasks: (agentId: string) => request<Task[]>(`/api/agents/${agentId}/tasks`),
  sendMessage: (agentId: string, message: string) =>
    request<{ task: Task }>(`/api/agents/${agentId}/chat`, {
      method: "POST",
      body: JSON.stringify({ message }),
    }),

  listBindings: () => request<ModelBinding[]>("/api/bindings"),
  createBinding: (input: BindingInput & { api_key: string }) =>
    request<ModelBinding>("/api/bindings", { method: "POST", body: JSON.stringify(input) }),
  updateBinding: (id: string, input: BindingInput & { api_key?: string }) =>
    request<ModelBinding>(`/api/bindings/${id}`, { method: "PATCH", body: JSON.stringify(input) }),
  deleteBinding: (id: string) => request<void>(`/api/bindings/${id}`, { method: "DELETE" }),

  listSchedules: (agentId: string) => request<Schedule[]>(`/api/agents/${agentId}/schedules`),
  createSchedule: (agentId: string, input: { cron_expr: string; input: string }) =>
    request<Schedule>(`/api/agents/${agentId}/schedules`, {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateSchedule: (id: string, input: { cron_expr: string; input: string; enabled: boolean }) =>
    request<Schedule>(`/api/schedules/${id}`, { method: "PATCH", body: JSON.stringify(input) }),
  deleteSchedule: (id: string) => request<void>(`/api/schedules/${id}`, { method: "DELETE" }),

  listWebhooks: (agentId: string) => request<Webhook[]>(`/api/agents/${agentId}/webhooks`),
  createWebhook: (agentId: string) =>
    request<Webhook>(`/api/agents/${agentId}/webhooks`, { method: "POST" }),
  deleteWebhook: (id: string) => request<void>(`/api/webhooks/${id}`, { method: "DELETE" }),

  listAllTasks: () => request<TaskWithAgent[]>("/api/tasks"),

  listMemory: (agentId: string) => request<MemoryEntry[]>(`/api/agents/${agentId}/memory`),
  createMemory: (agentId: string, input: { entry_type: MemoryEntryType; content: string }) =>
    request<MemoryEntry>(`/api/agents/${agentId}/memory`, {
      method: "POST",
      body: JSON.stringify(input),
    }),
  updateMemory: (id: string, input: { entry_type: MemoryEntryType; content: string }) =>
    request<MemoryEntry>(`/api/memory/${id}`, { method: "PATCH", body: JSON.stringify(input) }),
  deleteMemory: (id: string) => request<void>(`/api/memory/${id}`, { method: "DELETE" }),

  listRedLineRules: (agentId: string) => request<RedLineRule[]>(`/api/agents/${agentId}/redlines`),
  createRedLineRule: (agentId: string, pattern: string) =>
    request<RedLineRule>(`/api/agents/${agentId}/redlines`, {
      method: "POST",
      body: JSON.stringify({ pattern }),
    }),
  deleteRedLineRule: (id: string) => request<void>(`/api/redlines/${id}`, { method: "DELETE" }),

  approveTask: (taskId: string) =>
    request<{ task: Task }>(`/api/tasks/${taskId}/approve`, { method: "POST" }),
  rejectTask: (taskId: string) =>
    request<{ task: Task }>(`/api/tasks/${taskId}/reject`, { method: "POST" }),
};

type BindingInput = {
  provider: string;
  model_name: string;
  config?: Record<string, unknown>;
};
