package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/grassrootseconomics/pretium-ramp/internal/api"
	"github.com/grassrootseconomics/pretium-ramp/internal/kvvise"
	"github.com/grassrootseconomics/pretium-ramp/internal/store"
	"github.com/grassrootseconomics/pretium-ramp/internal/sub"
	"github.com/grassrootseconomics/pretium-ramp/internal/util"
	"github.com/grassrootseconomics/pretium-ramp/internal/worker"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 20

var (
	build = "dev"

	confFlag             string
	migrationsFolderFlag string
	queriesFlag          string

	lo *slog.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.StringVar(&migrationsFolderFlag, "migrations", "migrations/", "Migrations folder location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting pretium-ramp", "build", build)
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	store, err := store.NewPgStore(store.PgOpts{
		Logg:                 lo,
		DSN:                  ko.MustString("postgres.dsn"),
		MigrationsFolderPath: migrationsFolderFlag,
		QueriesFolderPath:    queriesFlag,
	})
	if err != nil {
		lo.Error("could not initialize postgres store", "error", err)
		os.Exit(1)
	}

	pretiumClient := pretium.New(
		ko.MustString("pretium.settlement_address"),
		ko.MustString("pretium.api_key"),
		ko.MustString("pretium.callback_url"),
	)

	kvviseClient := kvvise.New(
		ko.MustString("kvvise.endpoint"),
		ko.MustString("kvvise.api_key"),
	)

	workerContainer, err := worker.New(worker.WorkerOpts{
		MaxWorkers: runtime.NumCPU(),
		Store:      store,
		Logg:       lo,
		Pretium:    pretiumClient,
		KVVise:     kvviseClient,
	})
	if err != nil {
		lo.Error("could not initialize worker container", "error", err)
		os.Exit(1)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := workerContainer.Start(ctx); err != nil {
			lo.Error("worker container stopped", "error", err)
		}
	}()

	jetStreamSub, err := sub.NewJetStreamSub(sub.JetStreamOpts{
		Endpoint:    ko.MustString("jetstream.endpoint"),
		JetStreamID: ko.MustString("jetstream.consumer_id"),
		Logg:        lo,
		QueueClient: workerContainer.Client(),
	})
	if err != nil {
		lo.Error("could not initialize jetstream subscriber", "error", err)
		os.Exit(1)
	}

	apiServer := api.New(api.APIOpts{
		CallbackSecret: ko.MustString("api.callback_secret"),
		EnableMetrics:  ko.Bool("metrics.enable"),
		ListenAddress:  ko.MustString("api.address"),
		Logg:           lo,
		Store:          store,
		Pretium:        pretiumClient,
		KVVise:         kvviseClient,
		QueueClient:    workerContainer.Client(),
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetStreamSub.Process()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiServer.Start(); err != nil {
			lo.Error("failed to start HTTP server", "err", fmt.Sprintf("%T", err))
			os.Exit(1)
		}
	}()

	lo.Info("all services started successfully")

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetStreamSub.Close()
		if err := workerContainer.Stop(shutdownCtx); err != nil {
			lo.Error("failed to stop worker container", "error", err)
		}
		if err := apiServer.Stop(shutdownCtx); err != nil {
			lo.Error("failed to stop HTTP server", "err", fmt.Sprintf("%T", err))
		}
	}()

	go func() {
		wg.Wait()
		stop()
		cancel()
		os.Exit(0)
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		stop()
		cancel()
		lo.Error("graceful shutdown period exceeded, forcefully shutting down")
	}
	os.Exit(1)
}

func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}
