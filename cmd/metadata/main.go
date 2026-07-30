package main

import (
	"go-api/cmd/metadata/di"
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

	metadataWorker := rabbitmq.NewWorker(
		env,
		env.MetadataAnalyzeQueueName,
		container.MetadataHandler,
		env.WorkerConcurrency,
	)

	go func() {
		if err := metadataWorker.Start(); err != nil {
			log.Fatalf("failed to start metadata worker: %v", err)
		}
	}()

	log.Println("metadata started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh

	log.Printf("received signal %s, shutting down", sig)

	if err := metadataWorker.Stop(); err != nil {
		log.Printf("failed to stop metadata worker: %v", err)
	}
}
