package main

import (
	"go-api/cmd/aimodel/di"
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

	aiModelWorker := rabbitmq.NewWorker(
		env,
		env.AiModelAnalyzeQueueName,
		container.AiModelHandler,
		env.WorkerConcurrency,
	)

	go func() {
		if err := aiModelWorker.Start(); err != nil {
			log.Fatalf("failed to start ai_model worker: %v", err)
		}
	}()

	log.Println("ai_model started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh

	log.Printf("received signal %s, shutting down", sig)

	if err := aiModelWorker.Stop(); err != nil {
		log.Printf("failed to stop ai_model worker: %v", err)
	}
}
