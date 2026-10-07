package cron

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func Start() (gocron.Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}
	if _, err = scheduler.NewJob(
		gocron.DurationJob(time.Minute),
		gocron.NewTask(func() {
			slog.Info("cron heartbeat")
		}),
	); err != nil {
		return nil, fmt.Errorf("register heartbeat job: %w", err)
	}
	scheduler.Start()
	return scheduler, nil
}
