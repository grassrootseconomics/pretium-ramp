package api

import (
	"errors"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type LinkRequest struct {
	PublicKey   string `json:"publicKey" validate:"required,eth_addr_checksum"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

func (a *API) createLinkHandler(w http.ResponseWriter, req bunrouter.Request) error {
	var linkReq LinkRequest
	if err := httputil.BindJSON(w, req.Request, &linkReq); err != nil {
		a.logg.Error("failed to bind link request", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: "Invalid request body",
		})
	}

	if err := a.validator.Validate(linkReq); err != nil {
		a.logg.Error("validation failed", "error", err)
		return httputil.JSON(w, http.StatusBadRequest, ErrResponse{
			Ok:          false,
			Description: err.Error(),
		})
	}

	tx, err := a.store.Pool().Begin(req.Context())
	if err != nil {
		return handlePostgresError(w, err)
	}
	defer tx.Rollback(req.Context())

	// Check if this exact combination already exists (prevent duplicate links)
	existingLink, err := a.store.GetNonCustodialLinkByPublicKey(req.Context(), tx, linkReq.PublicKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to check existing link by public key", "error", err)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	// If this public key is already linked, check if it's to the same phone number
	if existingLink != nil && existingLink.Active {
		if existingLink.PhoneNumber == linkReq.PhoneNumber {
			// Already linked to the same phone number - this is idempotent
			a.logg.Info("link already exists with same details", "publicKey", linkReq.PublicKey, "phoneNumber", linkReq.PhoneNumber)
			return httputil.JSON(w, http.StatusOK, OKResponse{
				Ok:          true,
				Description: "Link already exists",
				Result: map[string]any{
					"publicKey":   linkReq.PublicKey,
					"phoneNumber": linkReq.PhoneNumber,
				},
			})
		} else {
			// Public key is linked to a different phone number. Re-link it:
			// deactivate the stale link scoped to this public key (so other
			// wallets sharing the old number keep their link) and insert the
			// new one below.
			a.logg.Info("re-linking public key to new phone", "publicKey", linkReq.PublicKey, "oldPhone", existingLink.PhoneNumber, "newPhone", linkReq.PhoneNumber)
			if err := a.store.DeactivateNonCustodialLinkByPublicKey(req.Context(), tx, linkReq.PublicKey); err != nil {
				a.logg.Error("failed to deactivate stale link", "error", err)
				return handlePostgresError(w, err)
			}
		}
	}

	// Note: We allow multiple public keys to be linked to the same phone number
	// This is by design to support users with multiple addresses

	err = a.store.InsertNonCustodialLink(req.Context(), tx, linkReq.PublicKey, linkReq.PhoneNumber)
	if err != nil {
		a.logg.Error("failed to insert non-custodial link", "error", err)
		return handlePostgresError(w, err)
	}

	if err := tx.Commit(req.Context()); err != nil {
		return handlePostgresError(w, err)
	}

	a.logg.Info("non-custodial link created", "publicKey", linkReq.PublicKey, "phoneNumber", linkReq.PhoneNumber)

	return httputil.JSON(w, http.StatusCreated, OKResponse{
		Ok:          true,
		Description: "Non-custodial link created successfully",
		Result: map[string]any{
			"publicKey":   linkReq.PublicKey,
			"phoneNumber": linkReq.PhoneNumber,
		},
	})
}

func (a *API) getLinksHandler(w http.ResponseWriter, req bunrouter.Request) error {
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

	links, err := a.store.GetNonCustodialLinksByPhone(req.Context(), tx, phoneNumber)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get links by phone", "error", err, "phoneNumber", phoneNumber)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	if err := tx.Commit(req.Context()); err != nil {
		return handlePostgresError(w, err)
	}

	a.logg.Info("links retrieved", "phoneNumber", phoneNumber, "count", len(links))

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Links retrieved successfully",
		Result: map[string]any{
			"phoneNumber": phoneNumber,
			"links":       links,
			"count":       len(links),
		},
	})
}

func (a *API) getLinkByAddressHandler(w http.ResponseWriter, req bunrouter.Request) error {
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

	link, err := a.store.GetNonCustodialLinkByPublicKey(req.Context(), tx, address)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		a.logg.Error("failed to get link by public key", "error", err, "address", address)
		return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
			Ok:          false,
			Description: "Internal server error",
		})
	}

	if err := tx.Commit(req.Context()); err != nil {
		return handlePostgresError(w, err)
	}

	if link == nil {
		a.logg.Info("no active link for address", "address", address)
		return httputil.JSON(w, http.StatusOK, OKResponse{
			Ok:          true,
			Description: "No active link for address",
			Result: map[string]any{
				"address": address,
				"linked":  false,
			},
		})
	}

	a.logg.Info("link retrieved", "address", address, "phoneNumber", link.PhoneNumber)

	return httputil.JSON(w, http.StatusOK, OKResponse{
		Ok:          true,
		Description: "Link retrieved successfully",
		Result: map[string]any{
			"address":     address,
			"linked":      true,
			"publicKey":   link.PublicKey,
			"phoneNumber": link.PhoneNumber,
		},
	})
}
