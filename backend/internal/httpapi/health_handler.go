package httpapi

import (
	"log/slog"
	"net/http"
)

type healthHandler struct {
	dependency HealthChecker
	logger     *slog.Logger
}

func newHealthHandler(dependency HealthChecker, logger *slog.Logger) *healthHandler {
	return &healthHandler{dependency: dependency, logger: logger}
}

type healthBody struct {
	Status string `json:"status"`
}

// check reports whether the API can reach its database, which is what an
// orchestrator polls to decide if this instance should receive traffic.
func (h *healthHandler) check(writer http.ResponseWriter, request *http.Request) {
	if err := h.dependency.Check(request.Context()); err != nil {
		h.logger.ErrorContext(request.Context(), "health check failed", "error", err)
		writeJSON(writer, h.logger, http.StatusServiceUnavailable, healthBody{Status: "unavailable"})
		return
	}
	writeJSON(writer, h.logger, http.StatusOK, healthBody{Status: "ok"})
}
