package api

import (
	"net/http"

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
