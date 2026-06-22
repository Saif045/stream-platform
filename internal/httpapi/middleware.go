package httpapi

import "net/http"

func requireHookSecret(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			writeError(w, http.StatusInternalServerError, "hook secret not configured")
			return
		}

		if r.Header.Get("X-Hook-Secret") != secret {
			writeError(w, http.StatusUnauthorized, "invalid hook secret")
			return
		}

		next.ServeHTTP(w, r)
	})
}
