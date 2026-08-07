package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/store"
)

func (a *API) listWebhooks(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	webhooks, err := a.store.ListWebhooksForAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, webhooks)
}

func (a *API) createWebhook(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	ctx := r.Context()

	if _, err := a.store.GetAgent(ctx, agentID); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := generateWebhookToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token: "+err.Error())
		return
	}

	created, err := a.store.CreateWebhook(ctx, models.Webhook{AgentID: agentID, Token: token})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) deleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	if err := a.store.DeleteWebhook(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "webhook not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// triggerWebhook is the public endpoint external services POST to. The token
// in the URL is the only auth — treat it like a bearer credential.
func (a *API) triggerWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	wh, err := a.store.GetWebhookByToken(ctx, token)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "webhook not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	agent, err := a.store.GetAgent(ctx, wh.AgentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	input := extractWebhookInput(body)
	if input == "" {
		writeError(w, http.StatusBadRequest, "request body must contain a non-empty message/input field, or raw text")
		return
	}

	task, err := a.runner.Run(ctx, agent, "webhook", input)
	if err != nil {
		writeRunError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Task: task})
}

// extractWebhookInput accepts either {"message": "..."} / {"input": "..."}
// JSON bodies (common for webhook senders) or plain text as a fallback.
func extractWebhookInput(body []byte) string {
	var parsed struct {
		Message string `json:"message"`
		Input   string `json:"input"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if parsed.Message != "" {
			return parsed.Message
		}
		if parsed.Input != "" {
			return parsed.Input
		}
	}
	return strings.TrimSpace(string(body))
}

func generateWebhookToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
