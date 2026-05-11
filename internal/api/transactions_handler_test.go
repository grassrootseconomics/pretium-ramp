package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestAPI builds an API with no store/pretium/kvvise. Sufficient for
// handler paths that return before any store call (e.g. input validation).
func newTestAPI(t *testing.T) *API {
	t.Helper()
	return New(APIOpts{
		Logg:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		ListenAddress: "127.0.0.1:0",
	})
}

func TestGetTransactionsByAddressHandler_ValidationErrors(t *testing.T) {
	tests := []struct {
		name            string
		address         string
		wantDescription string
	}{
		{
			name:            "rejects_non_hex_string",
			address:         "notanaddress",
			wantDescription: "Invalid Ethereum address",
		},
		{
			name:            "rejects_short_hex",
			address:         "0xabc",
			wantDescription: "Invalid Ethereum address",
		},
		{
			name:            "rejects_long_hex",
			address:         "0xcebA9300f2b948710d2653dD7B07f33A8B32118Cdeadbeef",
			wantDescription: "Invalid Ethereum address",
		},
		{
			name:            "rejects_non_hex_chars_in_address_shape",
			address:         "0xZZZZ9300f2b948710d2653dD7B07f33A8B32118C",
			wantDescription: "Invalid Ethereum address",
		},
	}

	api := newTestAPI(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions-by-address/"+tt.address, nil)
			rec := httptest.NewRecorder()
			api.router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d; want %d", rec.Code, http.StatusBadRequest)
			}

			var body ErrResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal response: %v (body: %s)", err, rec.Body.String())
			}
			if body.Ok {
				t.Errorf("body.Ok = true; want false")
			}
			if body.Description != tt.wantDescription {
				t.Errorf("description = %q; want %q", body.Description, tt.wantDescription)
			}
		})
	}
}

// Without a path segment, bunrouter does not match the route at all and
// returns the not-found handler (404). This guards that we never silently
// fall through to the empty-string code path.
func TestGetTransactionsByAddressHandler_MissingAddressIs404(t *testing.T) {
	api := newTestAPI(t)
	for _, path := range []string{
		"/api/v1/transactions-by-address",
		"/api/v1/transactions-by-address/",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		api.router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("path %s: status = %d; want %d (body: %s)",
				path, rec.Code, http.StatusNotFound, strings.TrimSpace(rec.Body.String()))
		}
	}
}
