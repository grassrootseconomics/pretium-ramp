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
	OfframpArgs struct {
		InitiatorAddress string `json:"initiatorAddress"`
		TransactionHash  string `json:"transactionHash"`
		Amount           string `json:"amount"`
		TokenAddress     string `json:"tokenAddress"`
	}

	OfframpWorker struct {
		river.WorkerDefaults[OfframpArgs]
		wc *WorkerContainer
	}
)

func (OfframpArgs) Kind() string { return "OFFRAMP" }

func (w *OfframpWorker) Work(ctx context.Context, job *river.Job[OfframpArgs]) error {
	tx, err := w.wc.store.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// TODO: Inject the actual merhcnat address here instead of hardcoding
	if job.Args.InitiatorAddress == "0x0000000000000000000000000000000000000000" || job.Args.InitiatorAddress == "0x8005ee53E57aB11E11eAA4EFe07Ee3835Dc02F98" {
		w.wc.logg.Info("this is an onramp request from pretium ignoreing offramp", "address", job.Args.InitiatorAddress)
		return nil
	}

	link, err := w.wc.store.GetNonCustodialLinkByPublicKey(ctx, tx, job.Args.InitiatorAddress)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check non_custodial_link: %w", err)
	}

	var phoneNumber string

	if link == nil || errors.Is(err, pgx.ErrNoRows) {
		w.wc.logg.Debug("address not found in non_custodial_link, checking kvvise reverse lookup", "address", job.Args.InitiatorAddress)
		resp, err := w.wc.kvvise.ReverseLookup(ctx, job.Args.InitiatorAddress)
		if err != nil {
			w.wc.logg.Warn("reverse lookup failed", "address", job.Args.InitiatorAddress, "error", err)
			return nil
		}

		phoneNumber = resp.Result.Phone
		w.wc.logg.Info("phone resolved from kvvise", "address", job.Args.InitiatorAddress, "phone", phoneNumber)
	} else {
		phoneNumber = link.PhoneNumber
		w.wc.logg.Info("phone resolved from database", "address", job.Args.InitiatorAddress, "phone", phoneNumber)
	}

	exchangeRateResp, err := w.wc.pretium.ExchangeRate(ctx, pretium.ExchangeRateBody{
		CurrencyCode: pretium.KES,
	})
	if err != nil {
		return fmt.Errorf("failed to get exchange rate: %w", err)
	}

	buyingRate := exchangeRateResp.Data.BuyingRate
	w.wc.logg.Debug("exchange rate received", "buyingRate", buyingRate)

	kesAmount, err := USDToKES(job.Args.Amount, job.Args.TokenAddress, buyingRate)
	if err != nil {
		return fmt.Errorf("failed to convert USD to KES: %w", err)
	}
	w.wc.logg.Debug("amount converted", "usd", job.Args.Amount, "kes", kesAmount)

	payResp, err := w.wc.pretium.Pay(ctx, pretium.PayBody{
		TransactionHash: job.Args.TransactionHash,
		Shortcode:       phoneNumber,
		Amount:          kesAmount,
		Type:            pretium.MOBILE,
		MobileNetwork:   pretium.SAFARICOM,
		Chain:           pretium.CELO,
	})
	if err != nil {
		w.wc.logg.Error("pretium pay error", "error", err, "transactionHash", job.Args.TransactionHash, "shortcode", phoneNumber, "amount", kesAmount)
		return nil
	}

	w.wc.logg.Info("pretium pay called", "transactionCode", payResp.Data.TransactionCode, "status", payResp.Data.Status)

	err = w.wc.store.InsertOfframp(
		ctx,
		tx,
		payResp.Data.TransactionCode,
		phoneNumber,
		job.Args.Amount,
		kesAmount,
		job.Args.TransactionHash,
		job.Args.TokenAddress,
	)
	if err != nil {
		w.wc.logg.Error("failed to insert offramp record", "error", err, "pretiumID", payResp.Data.TransactionCode)
		return nil
	}

	w.wc.logg.Debug("offramp record saved", "pretiumID", payResp.Data.TransactionCode)
	return tx.Commit(ctx)
}
