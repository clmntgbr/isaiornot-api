package rabbitmq

import (
	"log"
	"sync"

	"go-api/infrastructure/config"
)

type WorkerPool struct {
	workers []*Worker
}

func NewWorkerPool(workers ...*Worker) *WorkerPool {
	return &WorkerPool{workers: workers}
}

func (p *WorkerPool) Start() {
	var wg sync.WaitGroup

	for _, worker := range p.workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			if err := w.Start(); err != nil {
				log.Printf("worker %q stopped with error: %v", w.queueName, err)
			}
		}(worker)
	}

	wg.Wait()
}

func (p *WorkerPool) Stop() error {
	for _, worker := range p.workers {
		if err := worker.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func NewDispatcherWorkers(
	env *config.Config,
	analyzeRequestHandler MessageHandler,
	stageDoneHandler MessageHandler,
) *WorkerPool {
	return NewWorkerPool(
		NewWorker(env, env.AnalyzeRequestQueueName, analyzeRequestHandler, 1),
		NewWorker(env, env.MetadataDoneQueueName, stageDoneHandler, 1),
		NewWorker(env, env.HeuristicsDoneQueueName, stageDoneHandler, 1),
		NewWorker(env, env.AiModelDoneQueueName, stageDoneHandler, 1),
	)
}
