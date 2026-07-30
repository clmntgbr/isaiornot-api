package main

import (
	"go-api/cmd/scheduler/di"
	"go-api/cmd/scheduler/jobs"
	"go-api/infrastructure/config"
	"go-api/infrastructure/scheduler"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	env := config.Load()
	db := config.ConnectDatabase(env)

	container, err := di.NewContainer(db, env)
	if err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	sched := scheduler.New()
	if err := jobs.Register(sched, container.RetryStaleScansUseCase); err != nil {
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
