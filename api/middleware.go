package api

import (
	"log/slog"
	"net/http"
)

func Log(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		attrs := make([]slog.Attr, 0, 5)

		attrs = append(attrs, slog.Group("request", "Address", r.RemoteAddr, "method", r.Method, "url", r.URL))

		logger.LogAttrs(r.Context(), slog.LevelInfo, "Recieved Request", attrs...)

		next.ServeHTTP(w, r)
	})
}
