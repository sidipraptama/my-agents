// Package service holds the task execution logic shared by every trigger
// source (manual chat, API calls, schedules, webhooks).
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"myagents/internal/llm"
	"myagents/internal/models"
	"myagents/internal/persona"
	"myagents/internal/store"
)

var (
	ErrAgentDisabled  = errors.New("agent is paused")
	ErrNoModelBinding = errors.New("agent has no model binding configured")
)

type TaskRunner struct {
	store *store.Store
}

func NewTaskRunner(s *store.Store) *TaskRunner {
	return &TaskRunner{store: s}
}

// Run creates a task for the given agent and input, executes it against the
// agent's bound model, and persists the result — the single execution path
// used by manual chat, API-triggered tasks, schedules, and webhooks.
func (t *TaskRunner) Run(ctx context.Context, agent models.Agent, triggerType, input string) (models.Task, error) {
	if !agent.Enabled {
		return models.Task{}, ErrAgentDisabled
	}
	if agent.ModelBindingID == nil {
		return models.Task{}, ErrNoModelBinding
	}

	history, err := t.store.ListTasksForAgent(ctx, agent.ID)
	if err != nil {
		return models.Task{}, fmt.Errorf("list task history: %w", err)
	}

	task, err := t.store.CreateTask(ctx, models.Task{
		AgentID:     agent.ID,
		TriggerType: triggerType,
		Status:      models.TaskRunning,
		Input:       input,
		Transcript: []models.TranscriptMessage{
			{Role: "user", Content: input, Timestamp: time.Now()},
		},
	})
	if err != nil {
		return models.Task{}, fmt.Errorf("create task: %w", err)
	}

	provider, modelName, apiKey, config, err := t.store.ResolveCredentials(ctx, *agent.ModelBindingID)
	if err != nil {
		t.fail(ctx, task)
		return models.Task{}, fmt.Errorf("resolve model credentials: %w", err)
	}

	client, err := llm.New(provider, modelName, apiKey, config)
	if err != nil {
		t.fail(ctx, task)
		return models.Task{}, err
	}

	systemPrompt := persona.Compile(agent.Name, agent.Role, agent.Persona)
	messages := historyToMessages(history)
	messages = append(messages, llm.Message{Role: "user", Content: input})

	reply, err := client.Send(ctx, llm.Request{SystemPrompt: systemPrompt, Messages: messages})
	if err != nil {
		t.fail(ctx, task)
		return models.Task{}, fmt.Errorf("model call failed: %w", err)
	}

	task.Transcript = append(task.Transcript, models.TranscriptMessage{
		Role: "assistant", Content: reply, Timestamp: time.Now(),
	})
	task.Status = models.TaskDone
	if err := t.store.UpdateTaskResult(ctx, task.ID, task.Status, task.Transcript); err != nil {
		return models.Task{}, fmt.Errorf("persist task result: %w", err)
	}

	return task, nil
}

func (t *TaskRunner) fail(ctx context.Context, task models.Task) {
	_ = t.store.UpdateTaskResult(ctx, task.ID, models.TaskFailed, task.Transcript)
}

func historyToMessages(tasks []models.Task) []llm.Message {
	var out []llm.Message
	for _, tk := range tasks {
		for _, m := range tk.Transcript {
			out = append(out, llm.Message{Role: m.Role, Content: m.Content})
		}
	}
	return out
}
