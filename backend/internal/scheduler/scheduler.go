// Package scheduler runs recurring Agent tasks on a cron schedule, in-process.
package scheduler

import (
	"context"
	"log"
	"sync"

	"github.com/robfig/cron/v3"

	"myagents/internal/models"
	"myagents/internal/service"
	"myagents/internal/store"
)

type Scheduler struct {
	cron   *cron.Cron
	store  *store.Store
	runner *service.TaskRunner

	mu      sync.Mutex
	entries map[string]cron.EntryID // schedule ID -> cron entry ID
}

func New(s *store.Store, runner *service.TaskRunner) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		store:   s,
		runner:  runner,
		entries: map[string]cron.EntryID{},
	}
}

// Start loads every enabled schedule from the DB, registers it, and starts
// the cron loop. Call once at boot.
func (sch *Scheduler) Start(ctx context.Context) error {
	schedules, err := sch.store.ListEnabledSchedules(ctx)
	if err != nil {
		return err
	}
	for _, sc := range schedules {
		if err := sch.register(sc); err != nil {
			log.Printf("scheduler: skipping schedule %s (%q): %v", sc.ID, sc.CronExpr, err)
		}
	}
	sch.cron.Start()
	return nil
}

func (sch *Scheduler) Stop() {
	sch.cron.Stop()
}

// ValidateCronExpr checks a cron expression without registering anything —
// used to reject bad input at API creation time.
func ValidateCronExpr(expr string) error {
	_, err := cron.ParseStandard(expr)
	return err
}

// Sync (re)registers a schedule's cron entry to match its current definition
// and enabled state. Call any time a schedule is created or updated.
func (sch *Scheduler) Sync(sc models.Schedule) error {
	sch.mu.Lock()
	defer sch.mu.Unlock()

	sch.unregisterLocked(sc.ID)
	if !sc.Enabled {
		return nil
	}
	return sch.register(sc)
}

// Remove unregisters a schedule's cron entry — call when a schedule is deleted.
func (sch *Scheduler) Remove(scheduleID string) {
	sch.mu.Lock()
	defer sch.mu.Unlock()
	sch.unregisterLocked(scheduleID)
}

func (sch *Scheduler) register(sc models.Schedule) error {
	entryID, err := sch.cron.AddFunc(sc.CronExpr, func() { sch.run(sc) })
	if err != nil {
		return err
	}
	sch.entries[sc.ID] = entryID
	return nil
}

func (sch *Scheduler) unregisterLocked(scheduleID string) {
	if entryID, ok := sch.entries[scheduleID]; ok {
		sch.cron.Remove(entryID)
		delete(sch.entries, scheduleID)
	}
}

func (sch *Scheduler) run(sc models.Schedule) {
	ctx := context.Background()
	agent, err := sch.store.GetAgent(ctx, sc.AgentID)
	if err != nil {
		log.Printf("scheduler: schedule %s: failed to load agent: %v", sc.ID, err)
		return
	}
	if _, err := sch.runner.Run(ctx, agent, "scheduled", sc.Input); err != nil {
		log.Printf("scheduler: schedule %s (agent %s) failed: %v", sc.ID, agent.Name, err)
	}
}
