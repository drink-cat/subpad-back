package svc

import (
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/cron"
	"github.com/drink-cat/subpad-back/internal/eth"
	"github.com/drink-cat/subpad-back/internal/model"
)

type ServiceContext struct {
	Config    *config.Config
	DB        *gorm.DB
	Store     *model.Store
	Eth       *ethclient.Client
	Scheduler gocron.Scheduler
}

func New(cfg *config.Config) (*ServiceContext, error) {
	db, err := model.OpenMySQL(cfg.MySQL.DSN)
	if err != nil {
		return nil, err
	}
	if db == nil {
		slog.Info("mysql skipped: dsn is empty")
	} else if err = model.AutoMigrate(db); err != nil {
		closeDB(db)
		return nil, err
	}

	ethClient, err := eth.Dial(cfg.Eth.RPC)
	if err != nil {
		closeDB(db)
		return nil, err
	}
	if ethClient == nil {
		slog.Info("ethereum client skipped: rpc is empty")
	}

	var scheduler gocron.Scheduler
	if cfg.Cron.Enabled {
		scheduler, err = cron.Start()
		if err != nil {
			closeDB(db)
			if ethClient != nil {
				ethClient.Close()
			}
			return nil, err
		}
	}

	return &ServiceContext{
		Config:    cfg,
		DB:        db,
		Store:     model.NewStore(db),
		Eth:       ethClient,
		Scheduler: scheduler,
	}, nil
}

func (s *ServiceContext) Close() {
	if s.Scheduler != nil {
		if err := s.Scheduler.Shutdown(); err != nil {
			slog.Error("shutdown scheduler", "err", err)
		}
	}
	if s.Eth != nil {
		s.Eth.Close()
	}
	closeDB(s.DB)
}

func closeDB(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("get sql db", "err", err)
		return
	}
	if err = sqlDB.Close(); err != nil {
		slog.Error("close mysql", "err", err)
	}
}

func (s *ServiceContext) String() string {
	return fmt.Sprintf("mysql=%t eth=%t cron=%t", s.DB != nil, s.Eth != nil, s.Scheduler != nil)
}
