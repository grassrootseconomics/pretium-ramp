package api

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func notFoundHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusNotFound, ErrResponse{
		Ok:          false,
		Description: "Not found",
	})
}

func methodNotAllowedHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusMethodNotAllowed, ErrResponse{
		Ok:          false,
		Description: "Method not allowed",
	})
}

func handlePostgresError(w http.ResponseWriter, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return httputil.JSON(w, http.StatusNotFound, ErrResponse{
			Ok:          false,
			Description: "Record(s) not found",
		})
	}
	return httputil.JSON(w, http.StatusInternalServerError, ErrResponse{
		Ok:          false,
		Description: "Database error",
	})
}
