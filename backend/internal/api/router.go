package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"myagents/internal/scheduler"
	"myagents/internal/service"
	"myagents/internal/store"
)

type API struct {
	store         *store.Store
	runner        *service.TaskRunner
	scheduler     *scheduler.Scheduler
	allowedOrigin string
}

func New(s *store.Store, runner *service.TaskRunner, sch *scheduler.Scheduler, allowedOrigin string) *API {
	return &API{store: s, runner: runner, scheduler: sch, allowedOrigin: allowedOrigin}
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(a.cors)

	r.Route("/api", func(r chi.Router) {
		r.Route("/agents", func(r chi.Router) {
			r.Get("/", a.listAgents)
			r.Post("/", a.createAgent)
			r.Route("/{agentID}", func(r chi.Router) {
				r.Get("/", a.getAgent)
				r.Patch("/", a.updateAgent)
				r.Delete("/", a.deleteAgent)
				r.Get("/tasks", a.listTasks)
				r.Post("/tasks", a.createTask)
				r.Post("/chat", a.chat)
				r.Get("/schedules", a.listSchedules)
				r.Post("/schedules", a.createSchedule)
				r.Get("/webhooks", a.listWebhooks)
				r.Post("/webhooks", a.createWebhook)
				r.Get("/memory", a.listMemory)
				r.Post("/memory", a.createMemory)
				r.Get("/redlines", a.listRedLineRules)
				r.Post("/redlines", a.createRedLineRule)
			})
		})
		r.Route("/bindings", func(r chi.Router) {
			r.Get("/", a.listBindings)
			r.Post("/", a.createBinding)
			r.Route("/{bindingID}", func(r chi.Router) {
				r.Patch("/", a.updateBinding)
				r.Delete("/", a.deleteBinding)
			})
		})
		r.Route("/schedules/{scheduleID}", func(r chi.Router) {
			r.Patch("/", a.updateSchedule)
			r.Delete("/", a.deleteSchedule)
		})
		r.Delete("/webhooks/{webhookID}", a.deleteWebhook)
		r.Route("/memory/{memoryID}", func(r chi.Router) {
			r.Patch("/", a.updateMemory)
			r.Delete("/", a.deleteMemory)
		})
		r.Delete("/redlines/{redlineID}", a.deleteRedLineRule)
		r.Get("/tasks", a.listAllTasks)
		r.Route("/tasks/{taskID}", func(r chi.Router) {
			r.Post("/approve", a.approveTask)
			r.Post("/reject", a.rejectTask)
		})
		r.Post("/hooks/{token}", a.triggerWebhook)
	})

	return r
}

func (a *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", a.allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
