package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

func (s *Scheduler) Register(name, spec string, job func()) error {
	_, err := s.cron.AddFunc(spec, func() {
		startedAt := time.Now().UTC()
		log.Printf("cron job %q started (spec=%s)", name, spec)
		job()
		log.Printf("cron job %q finished in %s", name, time.Since(startedAt).Round(time.Millisecond))
	})
	if err != nil {
		return fmt.Errorf("failed to register cron job %q: %w", name, err)
	}
	return nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}
