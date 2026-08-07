export type Persona = {
  bio: string;
  tone: string;
  constraints: string[];
  knowledge: string[];
};

export type ModelBinding = {
  id: string;
  provider: string;
  model_name: string;
  config: Record<string, unknown>;
  created_at: string;
};

export type Agent = {
  id: string;
  name: string;
  role: string;
  persona: Persona;
  model_binding_id: string | null;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type TranscriptMessage = {
  role: "user" | "assistant";
  content: string;
  timestamp: string;
};

export type TaskStatus = "queued" | "running" | "done" | "failed";

export type Task = {
  id: string;
  agent_id: string;
  trigger_type: string;
  status: TaskStatus;
  input: string;
  transcript: TranscriptMessage[];
  created_at: string;
  updated_at: string;
};

export const emptyPersona: Persona = {
  bio: "",
  tone: "",
  constraints: [],
  knowledge: [],
};

export type TaskWithAgent = Task & { agent_name: string };

export type Schedule = {
  id: string;
  agent_id: string;
  cron_expr: string;
  input: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type Webhook = {
  id: string;
  agent_id: string;
  token: string;
  created_at: string;
};
