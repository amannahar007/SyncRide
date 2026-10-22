package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/syncride/backend/auth-service/internal/application/security"
	"github.com/syncride/backend/auth-service/internal/application/usecase"
	"github.com/syncride/backend/auth-service/internal/infrastructure/repository"
	myhttp "github.com/syncride/backend/auth-service/internal/interfaces/http"
	"github.com/syncride/backend/shared/config"
	"github.com/syncride/backend/shared/constants"
	"github.com/syncride/backend/shared/database"
	"github.com/syncride/backend/shared/jwt"
	"github.com/syncride/backend/shared/logger"
	"github.com/syncride/backend/shared/middleware"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("", constants.ServiceAuth)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	log, err := logger.InitLogger(constants.ServiceAuth, cfg.App.LogLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Auth Service", zap.String("version", cfg.App.Version))

	// 3. Initialize Database
	db, err := database.NewPostgresConnection(cfg.Database, log)
	if err != nil {
		log.Fatal("Database connection failed", zap.Error(err))
	}
	defer database.ClosePostgres(db)

	// 4. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// 5. Initialize Core Application Components
	tokenManager := jwt.NewTokenManager(cfg.JWT)
	hasher := security.NewPasswordHasher(nil) // uses default Argon2id params
	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, hasher, tokenManager, cfg.JWT.RefreshExpiry)

	// 6. Initialize HTTP Handlers
	authHandler := myhttp.NewAuthHandler(authUC)

	// 7. Setup Router & Global Middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.ZapLogger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS([]string{"*"}))
	r.Use(middleware.Gzip)

	// Health Check Endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := database.PingPostgres(r.Context(), db); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})

	// 8. Register Routes
	myhttp.RegisterRoutes(r, authHandler, tokenManager)

	// 9. Start HTTP Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Info(fmt.Sprintf("Auth Service listening on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.DrainTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Auth Service stopped")
}
