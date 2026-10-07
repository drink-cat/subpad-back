package cron

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/model"
	"github.com/drink-cat/subpad-back/internal/syncer"
)

func Start(store *model.Store, logs []config.SyncLogConfig) (gocron.Scheduler, func(), error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, nil, fmt.Errorf("create scheduler: %w", err)
	}
	chains := make([]*syncer.Chain, 0, len(logs))
	closeChains := func() {
		for _, chain := range chains {
			chain.Close()
		}
	}
	if store == nil {
		slog.Info("block scan skipped: mysql is not configured")
	} else {
		for _, item := range logs {
			if err = syncer.Valid(item); err != nil {
				slog.Warn("skip scan chain", "name", item.Name, "chainId", item.ChainID, "err", err)
				continue
			}
			chain := &syncer.Chain{Cfg: item, Store: store}
			chains = append(chains, chain)
			interval := time.Duration(item.TimerIntervalSeconds) * time.Second
			if interval <= 0 {
				interval = time.Minute
			}
			if _, err = scheduler.NewJob(
				gocron.DurationJob(interval),
				gocron.NewTask(func(ctx context.Context) {
					if err := chain.Run(ctx); err != nil {
						slog.Error("scan blocks", "name", chain.Cfg.Name, "chainId", chain.Cfg.ChainID, "err", err)
					}
				}),
				gocron.WithName(item.Name),
				gocron.WithStartAt(gocron.WithStartImmediately()),
				gocron.WithSingletonMode(gocron.LimitModeReschedule),
			); err != nil {
				closeChains()
				_ = scheduler.Shutdown()
				return nil, nil, fmt.Errorf("register scan job %s: %w", item.Name, err)
			}
			slog.Info("scan job started", "name", item.Name, "chainId", item.ChainID, "every", interval.String(), "blocks", syncer.BlocksPerScan)
		}
	}
	scheduler.Start()
	return scheduler, closeChains, nil
}
