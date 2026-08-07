package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

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
