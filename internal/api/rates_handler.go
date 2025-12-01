package api

import (
	"net/http"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func (a *API) getRatesHandler(w http.ResponseWriter, req bunrouter.Request) error {
	resp, err := a.pretium.ExchangeRate(req.Context(), pretium.ExchangeRateBody{
		CurrencyCode: pretium.KES,
	})
	if err != nil {
		a.logg.Error("failed to get rates", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Failed to get rates",
		})
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Rates retrieved successfully",
		Result: map[string]any{
			"buy":  resp.Data.BuyingRate,
			"sell": resp.Data.SellingRate,
		},
	})
}
