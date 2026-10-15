package httpapi

import (
	"log/slog"
	"net/http"
)

// LogRequestID runs inside identity assignment, before application handling.
// It records only the resolved, bounded context value, never caller headers,
// URLs, credentials, bodies or IPs. A nil logger explicitly disables the sink
// for embedded callers; production composition supplies its structured logger.
func LogRequestID(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if id := RequestID(request); logger != nil && id != "" {
			logger.InfoContext(request.Context(), "api request", "request_id", id)
		}
		next.ServeHTTP(response, request)
	})
}
