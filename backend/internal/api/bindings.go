package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/store"
)

func (a *API) listBindings(w http.ResponseWriter, r *http.Request) {
	bindings, err := a.store.ListBindings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, bindings)
}

func (a *API) createBinding(w http.ResponseWriter, r *http.Request) {
	var in models.ModelBinding
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Provider == "" || in.ModelName == "" || in.APIKey == "" {
		writeError(w, http.StatusBadRequest, "provider, model_name, and api_key are required")
		return
	}
	if msg := validateProviderConfig(in); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	created, err := a.store.CreateBinding(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) updateBinding(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bindingID")

	var in models.ModelBinding
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.Provider == "" || in.ModelName == "" {
		writeError(w, http.StatusBadRequest, "provider and model_name are required")
		return
	}
	if msg := validateProviderConfig(in); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	updated, err := a.store.UpdateBinding(r.Context(), id, in)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "binding not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) deleteBinding(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "bindingID")
	if err := a.store.DeleteBinding(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "binding not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateProviderConfig(b models.ModelBinding) string {
	if b.Provider == "litellm" {
		baseURL, _ := b.Config["base_url"].(string)
		if baseURL == "" {
			return `config.base_url is required for provider "litellm"`
		}
	}
	return ""
}
