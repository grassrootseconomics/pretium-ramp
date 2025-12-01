package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/grassrootseconomics/pretium-go"
	"github.com/grassrootseconomics/pretium-ramp/internal/kvvise"
	"github.com/grassrootseconomics/pretium-ramp/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/kamikazechaser/common/httputil"
	"github.com/riverqueue/river"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

type (
	APIOpts struct {
		CallbackSecret string
		EnableMetrics  bool
		ListenAddress  string
		Logg           *slog.Logger
		Store          store.Store
		Pretium        *pretium.PretiumClient
		KVVise         *kvvise.KVVise
		QueueClient    *river.Client[pgx.Tx]
	}

	API struct {
		validator   httputil.ValidatorProvider
		router      *bunrouter.Router
		server      *http.Server
		logg        *slog.Logger
		store       store.Store
		pretium     *pretium.PretiumClient
		kvvise      *kvvise.KVVise
		queueClient *river.Client[pgx.Tx]
	}
)

const apiVersion = "/api/v1"

func New(o APIOpts) *API {
	api := &API{
		validator:   httputil.NewValidator(""),
		logg:        o.Logg,
		store:       o.Store,
		pretium:     o.Pretium,
		kvvise:      o.KVVise,
		queueClient: o.QueueClient,
		router: bunrouter.New(
			bunrouter.WithNotFoundHandler(notFoundHandler),
			bunrouter.WithMethodNotAllowedHandler(methodNotAllowedHandler),
		),
	}

	if o.EnableMetrics {
		api.router.GET("/metrics", metricsHandler)
	}

	api.router.WithGroup(apiVersion, func(g *bunrouter.Group) {
		if os.Getenv("DEV") != "" {
			g = g.Use(reqlog.NewMiddleware())
		}

		g.POST(fmt.Sprintf("/callback/%s", o.CallbackSecret), api.callbackHandler)
		g.POST("/trigger-onramp", api.onrampHandler)
		g.POST("/link", api.createLinkHandler)
		g.GET("/link/:phoneNumber", api.getLinksHandler)
		g.GET("/transactions/:phoneNumber", api.getTransactionsByPhoneHandler)
		g.GET("/transactions-recent", api.getRecentTransactionsHandler)
		g.GET("/rates", api.getRatesHandler)
	})

	api.server = &http.Server{
		Addr:    o.ListenAddress,
		Handler: api.router,
	}

	return api
}

func (a *API) Start() error {
	a.logg.Info("API server starting", "address", a.server.Addr)
	if err := a.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *API) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
