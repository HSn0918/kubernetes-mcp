package middlewares

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	defaultAllowedHeaders = "Content-Type, Authorization, Mcp-Session-Id"
	defaultAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
)

func normalizeAllowedOrigins(allowOrigins string) string {
	trimmed := strings.TrimSpace(allowOrigins)
	if trimmed == "" {
		return "*"
	}
	return trimmed
}

func resolveAllowedOrigin(origin, allowOrigins string) (string, bool) {
	allowOrigins = normalizeAllowedOrigins(allowOrigins)
	if origin == "" {
		return "", true
	}
	if allowOrigins == "*" {
		return "*", true
	}
	for _, allowed := range strings.Split(allowOrigins, ",") {
		if strings.TrimSpace(allowed) == origin {
			return origin, true
		}
	}
	return "", false
}

func ApplyCorsHeaders(w http.ResponseWriter, r *http.Request, allowOrigins string) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	allowedOrigin, ok := resolveAllowedOrigin(origin, allowOrigins)
	if !ok {
		http.Error(w, fmt.Sprintf("origin %q is not allowed", origin), http.StatusForbidden)
		return true
	}

	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	}

	// Keep caches safe when origin/request headers vary.
	w.Header().Add("Vary", "Origin")
	w.Header().Add("Vary", "Access-Control-Request-Method")
	w.Header().Add("Vary", "Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Methods", defaultAllowedMethods)

	requestHeaders := strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers"))
	if requestHeaders == "" {
		requestHeaders = defaultAllowedHeaders
	}
	w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
	w.Header().Set("Access-Control-Max-Age", "600")

	// Browser forbids Allow-Credentials=true with Allow-Origin="*".
	if allowedOrigin != "" && allowedOrigin != "*" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		w.Header().Del("Access-Control-Allow-Credentials")
	}

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return true
	}

	return false
}

type corsProtectedResponseWriter struct {
	http.ResponseWriter
	header     http.Header
	wrote      bool
	lockedCORS map[string]string
}

func newCORSProtectedResponseWriter(w http.ResponseWriter) *corsProtectedResponseWriter {
	return &corsProtectedResponseWriter{
		ResponseWriter: w,
		header:         make(http.Header),
		lockedCORS:     make(map[string]string),
	}
}

func (w *corsProtectedResponseWriter) Header() http.Header {
	return w.header
}

func (w *corsProtectedResponseWriter) WriteHeader(statusCode int) {
	if w.wrote {
		return
	}
	w.wrote = true
	w.applyLockedCORSHeaders()
	dst := w.ResponseWriter.Header()
	for k := range dst {
		dst.Del(k)
	}
	for k, values := range w.header {
		dst[k] = append([]string(nil), values...)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *corsProtectedResponseWriter) Write(p []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *corsProtectedResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		if !w.wrote {
			w.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

func (w *corsProtectedResponseWriter) lockCORSHeaders() {
	corsHeaderKeys := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Credentials",
		"Access-Control-Max-Age",
	}
	for _, key := range corsHeaderKeys {
		w.lockedCORS[key] = w.header.Get(key)
	}
}

func (w *corsProtectedResponseWriter) applyLockedCORSHeaders() {
	for key, value := range w.lockedCORS {
		if value == "" {
			w.header.Del(key)
			continue
		}
		w.header.Set(key, value)
	}
}

func CorsMiddleware(allowOrigins string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsWriter := newCORSProtectedResponseWriter(w)
		if ApplyCorsHeaders(corsWriter, r, allowOrigins) {
			return
		}
		corsWriter.lockCORSHeaders()
		next.ServeHTTP(corsWriter, r)
	})
}

func CreateCorsHandlerFunc(allowOrigins string, defaultHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		corsWriter := newCORSProtectedResponseWriter(w)
		if ApplyCorsHeaders(corsWriter, r, allowOrigins) {
			return
		}
		corsWriter.lockCORSHeaders()
		defaultHandler.ServeHTTP(corsWriter, r)
	}
}
