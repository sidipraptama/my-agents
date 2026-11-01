package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/store"
)

func (a *API) listTasks(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	tasks, err := a.store.ListTasksForAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// listAllTasks powers the task board: every task across every Agent.
func (a *API) listAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := a.store.ListAllTasks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// createTask is the API-triggered entry point: any external caller can POST
// a message here to run it against an Agent, same as chat but for scripts
// instead of the console UI.
func (a *API) createTask(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")

	var in chatRequest
	if err := decodeJSON(r, &in); err != nil || in.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	ctx := r.Context()
	agent, err := a.store.GetAgent(ctx, agentID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	task, err := a.runner.Run(ctx, agent, "api", in.Message)
	if err != nil {
		writeRunError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Task: task})
}

func (a *API) approveTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	task, agent, ok := a.loadTaskAndAgent(w, r)
	if !ok {
		return
	}

	updated, err := a.runner.Approve(ctx, agent, task)
	if err != nil {
		writeRunError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Task: updated})
}

func (a *API) rejectTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	task, _, ok := a.loadTaskAndAgent(w, r)
	if !ok {
		return
	}

	updated, err := a.runner.Reject(ctx, task)
	if err != nil {
		writeRunError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Task: updated})
}

// loadTaskAndAgent fetches a task and its owning agent by the {taskID} URL
// param, writing the appropriate error response and returning ok=false on
// failure. Shared by approveTask and rejectTask.
func (a *API) loadTaskAndAgent(w http.ResponseWriter, r *http.Request) (task models.Task, agent models.Agent, ok bool) {
	id := chi.URLParam(r, "taskID")
	ctx := r.Context()

	task, err := a.store.GetTask(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return task, agent, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return task, agent, false
	}

	agent, err = a.store.GetAgent(ctx, task.AgentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return task, agent, false
	}

	return task, agent, true
}
