package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/syncride/backend/shared/jwt"
	"github.com/syncride/backend/shared/middleware"
)

// RegisterRoutes sets up the routing for the auth service
func RegisterRoutes(r chi.Router, handler *AuthHandler, tokenManager *jwt.TokenManager) {
	r.Route("/auth", func(r chi.Router) {
		// Public routes
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/refresh", handler.RefreshToken)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(tokenManager))
			
			r.Post("/logout", handler.Logout)
			r.Get("/me", handler.GetProfile)
		})
	})
}
