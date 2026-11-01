package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/store"
)

func (a *API) listMemory(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	entries, err := a.store.ListMemoryForAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (a *API) createMemory(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")

	var in models.MemoryEntry
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	in.AgentID = agentID

	created, err := a.store.CreateMemoryEntry(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) updateMemory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "memoryID")

	var in models.MemoryEntry
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	if in.EntryType == "" {
		in.EntryType = "fact"
	}

	updated, err := a.store.UpdateMemoryEntry(r.Context(), id, in.EntryType, in.Content)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "memory entry not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) deleteMemory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "memoryID")
	if err := a.store.DeleteMemoryEntry(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "memory entry not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
