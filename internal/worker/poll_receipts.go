package worker

import (
	"context"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/grassrootseconomics/pretium-ramp/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type (
	PollReceiptsArgs struct{}

	PollReceiptsWorker struct {
		river.WorkerDefaults[PollReceiptsArgs]
		wc *WorkerContainer
	}
)

const PollReceiptsID = "POLL_RECEIPTS"

func (PollReceiptsArgs) Kind() string { return PollReceiptsID }

func (w *PollReceiptsWorker) Work(ctx context.Context, _ *river.Job[PollReceiptsArgs]) error {
	tx, err := w.wc.store.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	staleOfframps, err := w.wc.store.GetStaleOfframps(ctx, tx)
	if err != nil {
		w.wc.logg.Debug("failed to get stale offramps", "error", err)
	} else {
		w.wc.logg.Debug("found stale offramps", "count", len(staleOfframps))
		for _, offramp := range staleOfframps {
			w.processStaleOfframp(ctx, tx, offramp)
		}
	}

	staleOnramps, err := w.wc.store.GetStaleOnramps(ctx, tx)
	if err != nil {
		w.wc.logg.Debug("failed to get stale onramps", "error", err)
	} else {
		w.wc.logg.Debug("found stale onramps", "count", len(staleOnramps))
		for _, onramp := range staleOnramps {
			w.processStaleOnramp(ctx, tx, onramp)
		}
	}

	return tx.Commit(ctx)
}

func (w *PollReceiptsWorker) processStaleOfframp(ctx context.Context, tx pgx.Tx, offramp store.Offramp) {
	statusResp, err := w.wc.pretium.Status(ctx, pretium.StatusBody{
		TransactionCode: offramp.PretiumID,
	})
	if err != nil {
		w.wc.logg.Debug("failed to fetch pretium status", "type", "offramp", "pretiumID", offramp.PretiumID, "error", err)
		return
	}

	err = w.wc.store.UpdateOfframpStatus(ctx, tx, statusResp.Data.Status, offramp.PretiumID)
	if err != nil {
		w.wc.logg.Debug("failed to update offramp status", "pretiumID", offramp.PretiumID, "error", err)
		return
	}

	if statusResp.Data.ReceiptNumber != nil && *statusResp.Data.ReceiptNumber != "" {
		err = w.wc.store.UpdateOfframpMpesaConfirmation(
			ctx,
			tx,
			*statusResp.Data.ReceiptNumber,
			statusResp.Data.Status,
			offramp.PretiumID,
		)
		if err != nil {
			w.wc.logg.Debug("failed to update mpesa confirmation", "type", "offramp", "pretiumID", offramp.PretiumID, "error", err)
			return
		}
		w.wc.logg.Info("updated stale offramp with mpesa confirmation", "pretiumID", offramp.PretiumID, "receiptNumber", *statusResp.Data.ReceiptNumber)
	} else {
		w.wc.logg.Debug("updated stale offramp status", "pretiumID", offramp.PretiumID, "status", statusResp.Data.Status)
	}
}

func (w *PollReceiptsWorker) processStaleOnramp(ctx context.Context, tx pgx.Tx, onramp store.Onramp) {
	statusResp, err := w.wc.pretium.Status(ctx, pretium.StatusBody{
		TransactionCode: onramp.PretiumID,
	})
	if err != nil {
		w.wc.logg.Debug("failed to fetch pretium status", "type", "onramp", "pretiumID", onramp.PretiumID, "error", err)
		return
	}

	err = w.wc.store.UpdateOnrampStatus(ctx, tx, statusResp.Data.Status, onramp.PretiumID)
	if err != nil {
		w.wc.logg.Debug("failed to update onramp status", "pretiumID", onramp.PretiumID, "error", err)
		return
	}

	if statusResp.Data.ReceiptNumber != nil && *statusResp.Data.ReceiptNumber != "" {
		err = w.wc.store.UpdateOnrampMpesaConfirmation(
			ctx,
			tx,
			*statusResp.Data.ReceiptNumber,
			statusResp.Data.Status,
			onramp.PretiumID,
		)
		if err != nil {
			w.wc.logg.Debug("failed to update mpesa confirmation", "type", "onramp", "pretiumID", onramp.PretiumID, "error", err)
			return
		}
		w.wc.logg.Info("updated stale onramp with mpesa confirmation", "pretiumID", onramp.PretiumID, "receiptNumber", *statusResp.Data.ReceiptNumber)
	} else {
		w.wc.logg.Debug("updated stale onramp status", "pretiumID", onramp.PretiumID, "status", statusResp.Data.Status)
	}
}
