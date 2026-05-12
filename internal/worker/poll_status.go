package worker

import (
	"context"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/grassrootseconomics/pretium-ramp/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type (
	PollStatusArgs struct{}

	PollStatusWorker struct {
		river.WorkerDefaults[PollStatusArgs]
		wc *WorkerContainer
	}
)

const PollStatusID = "POLL_STATUS"

func (PollStatusArgs) Kind() string { return PollStatusID }

func (w *PollStatusWorker) Work(ctx context.Context, _ *river.Job[PollStatusArgs]) error {
	tx, err := w.wc.store.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	pendingOfframps, err := w.wc.store.GetPendingOfframps(ctx, tx)
	if err != nil {
		w.wc.logg.Debug("failed to get pending offramps", "error", err)
	} else {
		w.wc.logg.Debug("found pending offramps", "count", len(pendingOfframps))
		for _, offramp := range pendingOfframps {
			w.processPendingOfframp(ctx, tx, offramp)
		}
	}

	pendingOnramps, err := w.wc.store.GetPendingOnramps(ctx, tx)
	if err != nil {
		w.wc.logg.Debug("failed to get pending onramps", "error", err)
	} else {
		w.wc.logg.Debug("found pending onramps", "count", len(pendingOnramps))
		for _, onramp := range pendingOnramps {
			w.processPendingOnramp(ctx, tx, onramp)
		}
	}

	return tx.Commit(ctx)
}

func (w *PollStatusWorker) processPendingOfframp(ctx context.Context, tx pgx.Tx, offramp store.Offramp) {
	statusResp, err := w.wc.pretium.Status(ctx, pretium.StatusBody{
		TransactionCode: offramp.PretiumID,
	})
	if err != nil {
		w.wc.logg.Debug("failed to fetch pretium status", "type", "offramp", "pretiumID", offramp.PretiumID, "error", err)
		return
	}

	if statusResp.Data.Status == offramp.PretiumStatus {
		w.wc.logg.Debug("offramp status unchanged, skipping update", "pretiumID", offramp.PretiumID, "status", offramp.PretiumStatus)
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
		w.wc.logg.Info("offramp status synced with mpesa confirmation", "pretiumID", offramp.PretiumID, "status", statusResp.Data.Status, "receiptNumber", *statusResp.Data.ReceiptNumber)
	} else {
		w.wc.logg.Info("offramp status synced", "pretiumID", offramp.PretiumID, "oldStatus", offramp.PretiumStatus, "newStatus", statusResp.Data.Status)
	}
}

func (w *PollStatusWorker) processPendingOnramp(ctx context.Context, tx pgx.Tx, onramp store.Onramp) {
	statusResp, err := w.wc.pretium.Status(ctx, pretium.StatusBody{
		TransactionCode: onramp.PretiumID,
	})
	if err != nil {
		w.wc.logg.Debug("failed to fetch pretium status", "type", "onramp", "pretiumID", onramp.PretiumID, "error", err)
		return
	}

	if statusResp.Data.Status == onramp.PretiumStatus {
		w.wc.logg.Debug("onramp status unchanged, skipping update", "pretiumID", onramp.PretiumID, "status", onramp.PretiumStatus)
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
		w.wc.logg.Info("onramp status synced with mpesa confirmation", "pretiumID", onramp.PretiumID, "status", statusResp.Data.Status, "receiptNumber", *statusResp.Data.ReceiptNumber)
	} else {
		w.wc.logg.Info("onramp status synced", "pretiumID", onramp.PretiumID, "oldStatus", onramp.PretiumStatus, "newStatus", statusResp.Data.Status)
	}
}
