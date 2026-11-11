package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
)

type api struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) http.Handler {
	a := &api{logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/operations/{operation}", a.calculate)
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("/", notFound)

	return withRequestID(withAccessLog(logger, withPanicRecovery(logger, mux)))
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, http.StatusNotFound, codeNotFound, fmt.Sprintf("path %q does not exist", r.URL.Path), "")
}
