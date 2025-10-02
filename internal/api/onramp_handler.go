package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type OnrampRequest struct {
	Address string  `json:"address" validate:"required,eth_addr_checksum"`
	Asset   string  `json:"asset" validate:"required,oneof=USDT USDC cUSD"`
	Amount  float64 `json:"amount" validate:"required,gte=20,lte=250000"`
}

func (a *API) onrampHandler(w http.ResponseWriter, req bunrouter.Request) error {
	var onrampReq OnrampRequest
	if err := httputil.BindJSON(w, req.Request, &onrampReq); err != nil {
		a.logg.Error("failed to bind onramp request", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Invalid request body",
		})
	}

	if err := a.validator.Validate(onrampReq); err != nil {
		a.logg.Error("validation failed", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: err.Error(),
		})
	}

	link, err := a.store.GetNonCustodialLinkByPublicKey(req.Context(), onrampReq.Address)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to check non_custodial_link", "error", err)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	var phoneNumber string

	if link == nil || errors.Is(err, pgx.ErrNoRows) {
		a.logg.Debug("address not found in non_custodial_link, checking kvvise", "address", onrampReq.Address)

		resp, err := a.kvvise.ReverseLookup(req.Context(), onrampReq.Address)
		if err != nil {
			a.logg.Warn("reverse lookup failed", "address", onrampReq.Address, "error", err)
			return httputil.JSON(w, http.StatusNotFound, ErrResponse{
				Ok:          false,
				Description: "Address not linked to any phone number",
			})
		}

		if resp.Result.Phone == "" {
			a.logg.Debug("no phone found in kvvise response", "address", onrampReq.Address)
			return httputil.JSON(w, http.StatusNotFound, ErrResponse{
				Ok:          false,
				Description: "No phone number found for address",
			})
		}

		phoneNumber = resp.Result.Phone
		a.logg.Info("phone resolved from kvvise", "address", onrampReq.Address, "phone", phoneNumber)
	} else {
		phoneNumber = link.PhoneNumber
		a.logg.Info("phone resolved from database", "address", onrampReq.Address, "phone", phoneNumber)
	}

	onrampResp, err := a.pretium.Onramp(req.Context(), pretium.KES, pretium.OnrampBody{
		Shortcode:     phoneNumber,
		Amount:        onrampReq.Amount,
		MobileNetwork: pretium.SAFARICOM,
		Chain:         pretium.CELO,
		Asset:         onrampReq.Asset,
		Address:       onrampReq.Address,
	})
	if err != nil {
		a.logg.Error("failed to trigger pretium onramp", "error", err)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Failed to trigger onramp",
		})
	}

	onrampID, err := a.store.InsertOnramp(
		req.Context(),
		onrampResp.Data.TransactionCode,
		phoneNumber,
		fmt.Sprintf("%.2f", onrampReq.Amount),
		fmt.Sprintf("%.2f", onrampReq.Amount),
		"",
		"",
	)
	if err != nil {
		a.logg.Error("failed to save onramp record", "error", err)
	} else {
		a.logg.Debug("onramp record saved", "id", onrampID)
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Onramp initiated successfully",
		Result: map[string]any{
			"transactionCode": onrampResp.Data.TransactionCode,
			"status":          onrampResp.Data.Status,
			"message":         onrampResp.Data.Message,
		},
	})
}
