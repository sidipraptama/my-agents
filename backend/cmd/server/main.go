package main

import (
	"context"
	"log"
	"net/http"

	"myagents/internal/api"
	"myagents/internal/config"
	"myagents/internal/crypto"
	"myagents/internal/db"
	"myagents/internal/scheduler"
	"myagents/internal/service"
	"myagents/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	conn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer conn.Close()

	if err := db.Migrate(conn, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	box, err := crypto.New(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("crypto: %v", err)
	}

	s := store.New(conn, box)
	runner := service.NewTaskRunner(s)

	sch := scheduler.New(s, runner)
	if err := sch.Start(context.Background()); err != nil {
		log.Fatalf("scheduler: %v", err)
	}
	defer sch.Stop()

	a := api.New(s, runner, sch, cfg.AllowedOrigin)

	log.Printf("my-agents backend listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, a.Router()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
