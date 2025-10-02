package api

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/uptrace/bunrouter"
)

func metricsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	metrics.WritePrometheus(w, true)
	return nil
}
