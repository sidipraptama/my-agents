package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/service"
	"myagents/internal/store"
)

type chatRequest struct {
	Message string `json:"message"`
}

type chatResponse struct {
	Task models.Task `json:"task"`
}

func (a *API) chat(w http.ResponseWriter, r *http.Request) {
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

	task, err := a.runner.Run(ctx, agent, "manual", in.Message)
	if err != nil {
		writeRunError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, chatResponse{Task: task})
}

// writeRunError maps a TaskRunner error to the right HTTP status. Shared by
// every handler that triggers a task run (chat, API tasks, webhooks).
func writeRunError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAgentDisabled),
		errors.Is(err, service.ErrNoModelBinding),
		errors.Is(err, service.ErrTaskNotWaitingApproval):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
}
