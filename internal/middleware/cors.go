package middleware

import (
	"net/http"
	"os"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		baseURL := os.Getenv("BASE_URL")
		allowed := false

		switch {
		case origin == baseURL:
			allowed = true
		case origin == "":
			allowed = true
		case len(origin) > 19 && origin[:19] == "chrome-extension://":
			allowed = true
		case len(origin) > 16 && origin[:16] == "moz-extension://":
			allowed = true
		}

		if allowed && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
