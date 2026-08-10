package jobs

import (
	"context"
	"log"

	scancmd "go-api/internal/application/command/scan"
	"go-api/internal/infrastructure/scheduler"
)

const retryStaleSpec = "@every 10m"

func Register(sched *scheduler.Scheduler, retryStale *scancmd.RetryStaleScansHandler) error {
	return sched.Register("retry-stale-scans", retryStaleSpec, func() {
		result, err := retryStale.Handle(context.Background())
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
