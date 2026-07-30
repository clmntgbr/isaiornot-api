package jobs

import (
	"context"
	"log"

	"go-api/infrastructure/scheduler"
	"go-api/usecase/scan"
)

const retryStaleSpec = "@every 10m"

func Register(sched *scheduler.Scheduler, retryStale *scan.RetryStaleScansUseCase) error {
	return sched.Register("retry-stale-scans", retryStaleSpec, func() {
		result, err := retryStale.Execute(context.Background())
		if err != nil {
			log.Printf("retry-stale-scans job failed: %v", err)
			return
		}
		log.Printf(
			"retry-stale-scans: %d retried, %d failed, %d media",
			result.Retried,
			result.Failed,
			result.Medias,
		)
	})
}
