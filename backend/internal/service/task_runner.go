// Package service holds the task execution logic shared by every trigger
// source (manual chat, API calls, schedules, webhooks).
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"myagents/internal/llm"
	"myagents/internal/models"
	"myagents/internal/persona"
	"myagents/internal/store"
)

var (
	ErrAgentDisabled          = errors.New("agent is paused")
	ErrNoModelBinding         = errors.New("agent has no model binding configured")
	ErrTaskNotWaitingApproval = errors.New("task is not waiting for approval")
)

const maxInjectedMemories = 20

type TaskRunner struct {
	store *store.Store
}

func NewTaskRunner(s *store.Store) *TaskRunner {
	return &TaskRunner{store: s}
}

// Run creates a task for the given agent and input. If the input matches one
// of the agent's red-line rules, the task is left in waiting_approval status
// and NOT executed — a human must call Approve or Reject. Otherwise it runs
// immediately. This is the single entry point used by manual chat,
// API-triggered tasks, schedules, and webhooks.
func (t *TaskRunner) Run(ctx context.Context, agent models.Agent, triggerType, input string) (models.Task, error) {
	if !agent.Enabled {
		return models.Task{}, ErrAgentDisabled
	}
	if agent.ModelBindingID == nil {
		return models.Task{}, ErrNoModelBinding
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

	matched, err := t.matchesRedLineRule(ctx, agent.ID, input)
	if err != nil {
		log.Printf("task runner: red-line rule check failed for agent %s: %v", agent.ID, err)
	}
	if matched {
		task.Status = models.TaskWaitingApproval
		if err := t.store.UpdateTaskResult(ctx, task.ID, task.Status, task.Transcript); err != nil {
			return models.Task{}, fmt.Errorf("persist waiting_approval status: %w", err)
		}
		return task, nil
	}

	return t.execute(ctx, agent, task)
}

// Approve resumes a task that was paused for approval, running it against the model.
func (t *TaskRunner) Approve(ctx context.Context, agent models.Agent, task models.Task) (models.Task, error) {
	if task.Status != models.TaskWaitingApproval {
		return models.Task{}, ErrTaskNotWaitingApproval
	}
	return t.execute(ctx, agent, task)
}

// Reject marks a paused task as rejected without ever running it against the model.
func (t *TaskRunner) Reject(ctx context.Context, task models.Task) (models.Task, error) {
	if task.Status != models.TaskWaitingApproval {
		return models.Task{}, ErrTaskNotWaitingApproval
	}
	task.Status = models.TaskRejected
	if err := t.store.UpdateTaskResult(ctx, task.ID, task.Status, task.Transcript); err != nil {
		return models.Task{}, fmt.Errorf("persist rejected status: %w", err)
	}
	return task, nil
}

// execute runs task.Input against agent's bound model and persists the
// result. Called either immediately from Run (no red-line match) or later
// from Approve, so it must not assume the task was just created.
func (t *TaskRunner) execute(ctx context.Context, agent models.Agent, task models.Task) (models.Task, error) {
	allHistory, err := t.store.ListTasksForAgent(ctx, agent.ID)
	if err != nil {
		t.fail(ctx, task)
		return models.Task{}, fmt.Errorf("list task history: %w", err)
	}
	history := make([]models.Task, 0, len(allHistory))
	for _, h := range allHistory {
		if h.ID != task.ID {
			history = append(history, h)
		}
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
	memories, err := t.store.ListMemoryForAgent(ctx, agent.ID)
	if err != nil {
		log.Printf("task runner: failed to load memory for agent %s: %v", agent.ID, err)
	} else {
		systemPrompt += formatMemories(memories)
	}

	messages := historyToMessages(history)
	messages = append(messages, llm.Message{Role: "user", Content: task.Input})

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

	if agent.AutoRemember {
		// Best-effort and fire-and-forget: capturing a memory shouldn't add
		// latency to the caller's response or fail the task if it errors.
		// Uses context.Background() since ctx is tied to the caller's request
		// and may be cancelled the moment execute returns.
		go t.captureMemory(context.Background(), agent, client, task.Input, reply)
	}

	return task, nil
}

func (t *TaskRunner) fail(ctx context.Context, task models.Task) {
	_ = t.store.UpdateTaskResult(ctx, task.ID, models.TaskFailed, task.Transcript)
}

func (t *TaskRunner) matchesRedLineRule(ctx context.Context, agentID, input string) (bool, error) {
	rules, err := t.store.ListRedLineRulesForAgent(ctx, agentID)
	if err != nil {
		return false, err
	}
	lowerInput := strings.ToLower(input)
	for _, rule := range rules {
		if rule.Pattern != "" && strings.Contains(lowerInput, strings.ToLower(rule.Pattern)) {
			return true, nil
		}
	}
	return false, nil
}

// captureMemory asks the model whether anything from this exchange is worth
// remembering long-term, and if so, saves it as a "summary" memory entry.
func (t *TaskRunner) captureMemory(ctx context.Context, agent models.Agent, client llm.Provider, userInput, assistantReply string) {
	prompt := fmt.Sprintf(
		"Based on this exchange, is there a durable fact, preference, or piece of context "+
			"worth remembering for future conversations?\n\nUser: %s\nAssistant: %s\n\n"+
			"If yes, reply with ONLY that memory as one concise sentence. "+
			"If nothing is worth remembering, reply with exactly: NONE",
		userInput, assistantReply,
	)

	result, err := client.Send(ctx, llm.Request{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		MaxTokens: 200,
	})
	if err != nil {
		log.Printf("task runner: memory capture failed for agent %s: %v", agent.ID, err)
		return
	}

	memory := strings.TrimSpace(result)
	if memory == "" || strings.EqualFold(memory, "NONE") {
		return
	}

	if _, err := t.store.CreateMemoryEntry(ctx, models.MemoryEntry{
		AgentID:   agent.ID,
		EntryType: "summary",
		Content:   memory,
	}); err != nil {
		log.Printf("task runner: failed to save captured memory for agent %s: %v", agent.ID, err)
	}
}

func formatMemories(entries []models.MemoryEntry) string {
	if len(entries) == 0 {
		return ""
	}
	if len(entries) > maxInjectedMemories {
		entries = entries[:maxInjectedMemories]
	}

	var b strings.Builder
	b.WriteString("\nWhat you remember from past interactions:\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "- %s\n", e.Content)
	}
	return b.String()
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
