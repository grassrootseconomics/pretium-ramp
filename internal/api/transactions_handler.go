package api

import (
	"errors"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func (a *API) getTransactionsByPhoneHandler(w http.ResponseWriter, req bunrouter.Request) error {
	phoneNumber := req.Param("phoneNumber")
	if phoneNumber == "" {
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Phone number is required",
		})
	}

	tx, err := a.store.Pool().Begin(req.Context())
	if err != nil {
		return handlePostgresError(w, err)
	}
	defer tx.Rollback(req.Context())

	onramps, err := a.store.GetOnrampsByPhone(req.Context(), tx, phoneNumber)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get onramps by phone", "error", err, "phoneNumber", phoneNumber)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	offramps, err := a.store.GetOfframpsByPhone(req.Context(), tx, phoneNumber)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get offramps by phone", "error", err, "phoneNumber", phoneNumber)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	if err := tx.Commit(req.Context()); err != nil {
		return handlePostgresError(w, err)
	}

	a.logg.Info("transactions retrieved", "phoneNumber", phoneNumber, "onrampCount", len(onramps), "offrampCount", len(offramps))

	result := map[string]any{
		"phoneNumber": phoneNumber,
		"onramps":     onramps,
		"offramps":    offramps,
		"totalCount":  len(onramps) + len(offramps),
	}

	if len(onramps) == 0 && len(offramps) == 0 {
		a.logg.Debug("no transactions found for phone number", "phoneNumber", phoneNumber)
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Transactions retrieved successfully",
		Result:      result,
	})
}

func (a *API) getTransactionsByAddressHandler(w http.ResponseWriter, req bunrouter.Request) error {
	address := req.Param("address")
	if address == "" {
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Address is required",
		})
	}
	if !common.IsHexAddress(address) {
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Invalid Ethereum address",
		})
	}
	address = common.HexToAddress(address).Hex()

	tx, err := a.store.Pool().Begin(req.Context())
	if err != nil {
		return handlePostgresError(w, err)
	}
	defer tx.Rollback(req.Context())

	onramps, err := a.store.GetOnrampsByWalletAddress(req.Context(), tx, address)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get onramps by wallet address", "error", err, "address", address)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	offramps, err := a.store.GetOfframpsByWalletAddress(req.Context(), tx, address)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get offramps by wallet address", "error", err, "address", address)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	if err := tx.Commit(req.Context()); err != nil {
		return handlePostgresError(w, err)
	}

	a.logg.Info("transactions retrieved", "address", address, "onrampCount", len(onramps), "offrampCount", len(offramps))

	result := map[string]any{
		"address":    address,
		"onramps":    onramps,
		"offramps":   offramps,
		"totalCount": len(onramps) + len(offramps),
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Transactions retrieved successfully",
		Result:      result,
	})
}

func (a *API) getRecentTransactionsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	tx, err := a.store.Pool().Begin(req.Context())
	if err != nil {
		return handlePostgresError(w, err)
	}
	defer tx.Rollback(req.Context())

	onramps, err := a.store.GetRecentOnramps(req.Context(), tx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get recent onramps", "error", err)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	offramps, err := a.store.GetRecentOfframps(req.Context(), tx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get recent offramps", "error", err)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	if err := tx.Commit(req.Context()); err != nil {
		return handlePostgresError(w, err)
	}

	a.logg.Info("recent transactions retrieved", "onrampCount", len(onramps), "offrampCount", len(offramps))

	result := map[string]any{
		"onramps":    onramps,
		"offramps":   offramps,
		"totalCount": len(onramps) + len(offramps),
		"period":     "last 3 days",
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Recent transactions retrieved successfully",
		Result:      result,
	})
}
