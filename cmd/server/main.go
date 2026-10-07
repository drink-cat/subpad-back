package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/handler"
	"github.com/drink-cat/subpad-back/internal/svc"
)

func main() {
	configFile := flag.String("f", "etc/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	sc, err := svc.New(cfg)
	if err != nil {
		slog.Error("init service", "err", err)
		os.Exit(1)
	}
	defer sc.Close()

	srv := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: handler.NewRouter(sc),
	}

	go func() {
		slog.Info("server listening", "addr", srv.Addr, "deps", sc.String())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown server", "err", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
