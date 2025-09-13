package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"subs-collector/config"
	"subs-collector/internal/handler"
	"subs-collector/internal/logger"
	"subs-collector/internal/repository"
	"subs-collector/internal/service"
)

func main() {
	l := logger.New()
	l.Info("start app")

	cfg := config.Load(".env", "config.yaml")
	l.Info("load configuration", "port", cfg.Port)

	server, dbPoolClose := func() (*http.Server, func()) {
		ctx := context.Background()
		pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			l.Error("error create datbase pool", "err", err)
			os.Exit(1)
		}

		if err := pool.Ping(ctx); err != nil {
			l.Error("failed ping to database", "err", err)
			os.Exit(1)
		}
		l.Info("connected to database")

		repo := repository.NewSubscriptionRepository(pool)
		svc := service.NewSubscriptionService(repo)
		h := handler.NewSubscriptionHandler(svc, l)

		mux := http.NewServeMux()
		h.Register(mux)
		wrapped := handler.CORS(mux)

		return &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           wrapped,
			ReadHeaderTimeout: 5 * time.Second,
		}, pool.Close
	}()
	defer dbPoolClose()

	go func() {
		l.Info("run http server listener", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("fatal http server", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	l.Info("stop app")
}
