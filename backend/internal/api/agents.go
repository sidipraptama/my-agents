package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/store"
)

func (a *API) listAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := a.store.ListAgents(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, agents)
}

func (a *API) createAgent(w http.ResponseWriter, r *http.Request) {
	var in models.Agent
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	in.Enabled = true

	created, err := a.store.CreateAgent(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) getAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "agentID")
	agent, err := a.store.GetAgent(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (a *API) updateAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "agentID")

	existing, err := a.store.GetAgent(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Start from the existing record so fields the client omits (e.g. the
	// model-binding picker only sends model_binding_id) keep their current
	// value instead of being zeroed out by the JSON decode.
	in := existing
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.ID = id

	updated, err := a.store.UpdateAgent(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) deleteAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "agentID")
	if err := a.store.DeleteAgent(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
