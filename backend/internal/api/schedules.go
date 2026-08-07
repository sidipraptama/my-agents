package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/models"
	"myagents/internal/scheduler"
	"myagents/internal/store"
)

func (a *API) listSchedules(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	schedules, err := a.store.ListSchedulesForAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, schedules)
}

func (a *API) createSchedule(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")

	var in models.Schedule
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.CronExpr == "" || in.Input == "" {
		writeError(w, http.StatusBadRequest, "cron_expr and input are required")
		return
	}
	if err := scheduler.ValidateCronExpr(in.CronExpr); err != nil {
		writeError(w, http.StatusBadRequest, "invalid cron_expr: "+err.Error())
		return
	}

	ctx := r.Context()
	if _, err := a.store.GetAgent(ctx, agentID); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	in.AgentID = agentID
	in.Enabled = true

	created, err := a.store.CreateSchedule(ctx, in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := a.scheduler.Sync(created); err != nil {
		writeError(w, http.StatusInternalServerError, "schedule saved but failed to register: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) updateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "scheduleID")

	existing, err := a.store.GetSchedule(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "schedule not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	in := existing
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if in.CronExpr == "" || in.Input == "" {
		writeError(w, http.StatusBadRequest, "cron_expr and input are required")
		return
	}
	if err := scheduler.ValidateCronExpr(in.CronExpr); err != nil {
		writeError(w, http.StatusBadRequest, "invalid cron_expr: "+err.Error())
		return
	}

	updated, err := a.store.UpdateSchedule(r.Context(), id, in)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "schedule not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := a.scheduler.Sync(updated); err != nil {
		writeError(w, http.StatusInternalServerError, "schedule saved but failed to register: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "scheduleID")
	if err := a.store.DeleteSchedule(r.Context(), id); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "schedule not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.scheduler.Remove(id)
	w.WriteHeader(http.StatusNoContent)
}
