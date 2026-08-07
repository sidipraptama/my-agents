package models

import "time"

type Persona struct {
	Bio         string   `json:"bio"`
	Tone        string   `json:"tone"`
	Constraints []string `json:"constraints"`
	Knowledge   []string `json:"knowledge"`
}

type ModelBinding struct {
	ID        string         `json:"id"`
	Provider  string         `json:"provider"`
	ModelName string         `json:"model_name"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
	// APIKey is only ever populated on create/update requests, never on read.
	APIKey string `json:"api_key,omitempty"`
}

type Agent struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	Persona        Persona   `json:"persona"`
	ModelBindingID *string   `json:"model_binding_id"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TaskStatus string

const (
	TaskQueued  TaskStatus = "queued"
	TaskRunning TaskStatus = "running"
	TaskDone    TaskStatus = "done"
	TaskFailed  TaskStatus = "failed"
)

type TranscriptMessage struct {
	Role      string    `json:"role"` // "user" | "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Task struct {
	ID          string              `json:"id"`
	AgentID     string              `json:"agent_id"`
	TriggerType string              `json:"trigger_type"`
	Status      TaskStatus          `json:"status"`
	Input       string              `json:"input"`
	Transcript  []TranscriptMessage `json:"transcript"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// TaskWithAgent is Task plus the owning Agent's name, used by the task board
// so it doesn't have to issue a separate lookup per row.
type TaskWithAgent struct {
	Task
	AgentName string `json:"agent_name"`
}

type Schedule struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	CronExpr  string    `json:"cron_expr"`
	Input     string    `json:"input"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Webhook struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}
