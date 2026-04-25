package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manoskammas/finance-insights/apps/api/internal/config"
	"github.com/manoskammas/finance-insights/apps/api/internal/db"
	"github.com/manoskammas/finance-insights/apps/api/internal/parser"
	"github.com/manoskammas/finance-insights/apps/api/internal/repository"
	"github.com/manoskammas/finance-insights/apps/api/internal/server"
	"github.com/manoskammas/finance-insights/apps/api/internal/service"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Error("db migrate", "err", err)
		os.Exit(1)
	}

	transactionRepo := repository.NewTransactionRepository(pool)
	merchantRepo := repository.NewMerchantRepository(pool)
	reportRepo := repository.NewReportRepository(pool)

	sys := service.NewSystem()
	statementSvc := service.NewStatement(pool, parser.NewParser(), transactionRepo, cfg.StorageDir)
	transactionSvc := service.NewTransaction(transactionRepo)
	merchantSvc := service.NewMerchant(merchantRepo)
	reportSvc := service.NewReport(reportRepo)

	handler := server.NewRouter(server.Deps{
		System:      sys,
		Statement:   statementSvc,
		Transaction: transactionSvc,
		Merchant:    merchantSvc,
		Report:      reportSvc,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "err", err)
		os.Exit(1)
	}
}
