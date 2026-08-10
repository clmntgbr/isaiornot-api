package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-api/cmd/scheduler/di"
	"go-api/cmd/scheduler/jobs"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/schema"
	"go-api/internal/infrastructure/scheduler"
)

func main() {
	env := config.Load()
	db := config.ConnectDatabase(env)

	if err := schema.AssertModelsMatchDB(db); err != nil {
		log.Fatalf("schema check failed: %v", err)
	}

	container := di.NewContainer(db, env)

	sched := scheduler.New()
	if err := jobs.Register(sched, container.RetryStaleScansHandler); err != nil {
		log.Fatalf("failed to register jobs: %v", err)
	}

	sched.Start()
	log.Println("scheduler started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Printf("received signal %s, shutting down", sig)
	sched.Stop()
}
