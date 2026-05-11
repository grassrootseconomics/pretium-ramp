package api

import (
	"net/http"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/grassrootseconomics/pretium-ramp/internal/worker"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func (a *API) callbackHandler(w http.ResponseWriter, req bunrouter.Request) error {
	payload, err := pretium.ParseWebhook(req.Body)
	if err != nil {
		a.logg.Error("failed to read callback payload", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Failed to read request body",
		})
	}
	a.logg.Info("received callback payload", "transactionCode", payload.TransactionCode, "status", payload.Status, "message", payload.Message)

	_, err = a.queueClient.Insert(req.Context(), worker.CallbackArgs{
		Payload: payload,
	}, nil)
	if err != nil {
		a.logg.Error("failed to queue callback job", "error", err)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Failed to queue callback processing",
		})
	}

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Callback received",
	})
}
