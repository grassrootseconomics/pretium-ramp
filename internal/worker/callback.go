package worker

import (
	"context"
	"errors"
	"fmt"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type (
	CallbackArgs struct {
		Payload pretium.WebhookPayload `json:"payload"`
	}

	CallbackWorker struct {
		river.WorkerDefaults[CallbackArgs]
		wc *WorkerContainer
	}
)

func (CallbackArgs) Kind() string { return "CALLBACK" }

func (w *CallbackWorker) Work(ctx context.Context, job *river.Job[CallbackArgs]) error {
	webhook := job.Args.Payload
	err := w.wc.store.UpdateOfframpStatus(ctx, webhook.Status, webhook.TransactionCode)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to update offramp status: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		err = w.wc.store.UpdateOnrampStatus(ctx, webhook.Status, webhook.TransactionCode)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to update onramp status: %w", err)
		}

		if errors.Is(err, pgx.ErrNoRows) {
			w.wc.logg.Warn("transaction code not found in either offramp or onramp", "transactionCode", webhook.TransactionCode)
			return nil
		}
	}

	w.wc.logg.Debug("status updated successfully", "transactionCode", webhook.TransactionCode, "status", webhook.Status)

	statusResp, err := w.wc.pretium.Status(ctx, pretium.StatusBody{
		TransactionCode: webhook.TransactionCode,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch pretium status: %w", err)
	}

	if statusResp.Data.ReceiptNumber != nil && *statusResp.Data.ReceiptNumber != "" {
		offramp, err := w.wc.store.GetOfframpByPretiumID(ctx, webhook.TransactionCode)
		if err == nil && offramp != nil {
			err = w.wc.store.UpdateOfframpMpesaConfirmation(
				ctx,
				*statusResp.Data.ReceiptNumber,
				statusResp.Data.Status,
				offramp.ID,
			)
			if err != nil {
				return fmt.Errorf("failed to update offramp mpesa confirmation: %w", err)
			}
			w.wc.logg.Info("offramp mpesa confirmation updated", "id", offramp.ID, "receiptNumber", *statusResp.Data.ReceiptNumber)
			return nil
		}

		onramp, err := w.wc.store.GetOnrampByPretiumID(ctx, webhook.TransactionCode)
		if err == nil && onramp != nil {
			err = w.wc.store.UpdateOnrampMpesaConfirmation(
				ctx,
				*statusResp.Data.ReceiptNumber,
				statusResp.Data.Status,
				onramp.ID,
			)
			if err != nil {
				return fmt.Errorf("failed to update onramp mpesa confirmation: %w", err)
			}
			w.wc.logg.Info("onramp mpesa confirmation updated", "id", onramp.ID, "receiptNumber", *statusResp.Data.ReceiptNumber)
			return nil
		}

		w.wc.logg.Warn("could not find transaction to update mpesa confirmation", "transactionCode", webhook.TransactionCode)
	} else {
		w.wc.logg.Debug("no receipt number in status response", "transactionCode", webhook.TransactionCode)
	}

	return nil
}
