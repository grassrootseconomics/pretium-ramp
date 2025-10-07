package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/grassrootseconomics/pretium-ramp/internal/kvvise"
	"github.com/grassrootseconomics/pretium-ramp/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

type (
	WorkerOpts struct {
		MaxWorkers int
		Store      store.Store
		Logg       *slog.Logger
		Pretium    *pretium.PretiumClient
		KVVise     *kvvise.KVVise
	}

	WorkerContainer struct {
		queueClient *river.Client[pgx.Tx]
		store       store.Store
		logg        *slog.Logger
		pretium     *pretium.PretiumClient
		kvvise      *kvvise.KVVise
	}
)

const (
	migrationTimeout = 15 * time.Second
	pollInterval     = 2 * time.Minute
)

func New(o WorkerOpts) (*WorkerContainer, error) {
	workerContainer := &WorkerContainer{
		queueClient: nil,
		store:       o.Store,
		logg:        o.Logg,
		pretium:     o.Pretium,
		kvvise:      o.KVVise,
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()

	riverPgxDriver := riverpgxv5.New(o.Store.Pool())
	riverMigrator, err := rivermigrate.New(riverPgxDriver, &rivermigrate.Config{
		Logger: o.Logg,
	})
	if err != nil {
		return nil, err
	}

	_, err = riverMigrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return nil, err
	}

	workers, err := setupWorkers(workerContainer)
	if err != nil {
		return nil, err
	}

	workerContainer.queueClient, err = river.NewClient(riverPgxDriver, &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {
				MaxWorkers: o.MaxWorkers,
			},
		},
		Workers:      workers,
		PeriodicJobs: setupPollReceiptsCheck(),
		Logger:       o.Logg,
	})
	if err != nil {
		return nil, err
	}

	return workerContainer, nil
}

func (w *WorkerContainer) Start(ctx context.Context) error {
	return w.queueClient.Start(ctx)
}

func (w *WorkerContainer) Stop(ctx context.Context) error {
	w.logg.Info("shutting down river queue")
	return w.queueClient.Stop(ctx)
}

func (w *WorkerContainer) Client() *river.Client[pgx.Tx] {
	return w.queueClient
}

func setupWorkers(wc *WorkerContainer) (*river.Workers, error) {
	workers := river.NewWorkers()

	river.AddWorker(workers, &PollReceiptsWorker{wc: wc})
	river.AddWorker(workers, &OfframpWorker{wc: wc})
	river.AddWorker(workers, &CallbackWorker{wc: wc})

	return workers, nil
}

func setupPollReceiptsCheck() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(pollInterval),
			func() (river.JobArgs, *river.InsertOpts) {
				return PollReceiptsArgs{}, nil
			},
			&river.PeriodicJobOpts{
				RunOnStart: true,
			},
		),
	}
}
