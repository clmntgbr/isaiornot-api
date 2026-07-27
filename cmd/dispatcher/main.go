package main

import (
	"go-api/cmd/dispatcher/di"
	"go-api/infrastructure/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	env := config.Load()
	db := config.ConnectDatabase(env)

	container := di.NewContainer(db, env)

	go container.WorkerPool.Start()

	log.Println("dispatcher started")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh

	log.Printf("received signal %s, shutting down", sig)

	if err := container.WorkerPool.Stop(); err != nil {
		log.Printf("failed to stop dispatcher workers: %v", err)
	}
}
