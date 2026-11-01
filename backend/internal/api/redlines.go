package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/store"
)

func (a *API) listRedLineRules(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	rules, err := a.store.ListRedLineRulesForAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (a *API) createRedLineRule(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")

	var in models.RedLineRule
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Pattern == "" {
		writeError(w, http.StatusBadRequest, "pattern is required")
		return
	}
	in.AgentID = agentID

	created, err := a.store.CreateRedLineRule(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) deleteRedLineRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "redlineID")
	if err := a.store.DeleteRedLineRule(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "red line rule not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
