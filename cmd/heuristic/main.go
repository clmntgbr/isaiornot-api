package main

import (
	"go-api/cmd/heuristic/di"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	env := config.Load()
	db := config.ConnectDatabase(env)

	container := di.NewContainer(db, env)

	heuristicWorker := rabbitmq.NewWorker(
		env,
		env.HeuristicsAnalyzeQueueName,
		container.HeuristicHandler,
		env.WorkerConcurrency,
	)

	go func() {
		if err := heuristicWorker.Start(); err != nil {
			log.Fatalf("failed to start heuristic worker: %v", err)
		}
	}()

	log.Println("heuristic started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh

	log.Printf("received signal %s, shutting down", sig)

	if err := heuristicWorker.Stop(); err != nil {
		log.Printf("failed to stop heuristic worker: %v", err)
	}
}
