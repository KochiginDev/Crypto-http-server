package httpapi

import (
	"crypthttp/usecases"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *AuthHandler, ch *CryptoHandler, auth usecases.Auth) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
	r.Route("/crypto", func(r chi.Router) {
		r.Use(AuthMiddleware(auth))

		r.Get("/", ch.GetList)
		r.Post("/", ch.Add)
		r.Get("/{symbol}", ch.Get)
		r.Delete("/{symbol}", ch.Delete)
		r.Put("/{symbol}/refresh", ch.Refresh)
		r.Get("/{symbol}/history", ch.GetHistory)
		r.Get("/{symbol}/stats", ch.GetStats)
	})
	return r
}
