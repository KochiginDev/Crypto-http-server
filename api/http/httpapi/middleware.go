package httpapi

import (
	"crypthttp/domain"
	"crypthttp/usecases"
	"net/http"
	"strings"
)

func AuthMiddleware(auth usecases.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok {
				ErrorWriter(w, r, domain.ErrInvalidCredentials)
				return
			}
			_, err := auth.CheckToken(r.Context(), token)
			if err != nil {
				ErrorWriter(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
