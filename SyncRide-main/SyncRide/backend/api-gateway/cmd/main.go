package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/syncride/backend/shared/config"
	"github.com/syncride/backend/shared/constants"
	"github.com/syncride/backend/shared/database"
	"github.com/syncride/backend/shared/logger"
	"github.com/syncride/backend/shared/middleware"
	"github.com/syncride/backend/shared/utils"
)

var (
	buildCommit  = "unknown"
	buildVersion = "1.0.0"
	buildTime    = "unknown"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("", constants.ServiceAPIGateway)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	log, err := logger.InitLogger(constants.ServiceAPIGateway, cfg.App.LogLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()
	log.Info("Starting API Gateway", zap.String("env", string(cfg.App.Environment)), zap.String("version", buildVersion))

	// 3. Initialize Database & Redis (Optional for Gateway, but verifying connections)
	db, err := database.NewPostgresConnection(cfg.Database, log)
	if err != nil {
		log.Error("Failed to connect to PostgreSQL", zap.Error(err))
	} else {
		defer database.ClosePostgres(db)
	}

	rdb, err := database.NewRedisClient(cfg.Redis, log)
	if err != nil {
		log.Error("Failed to connect to Redis", zap.Error(err))
	} else {
		defer database.CloseRedis(rdb)
	}

	// 4. Setup Routing & Middleware
	r := chi.NewRouter()

	// Base middlewares
	r.Use(middleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.ZapLogger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Gzip)
	r.Use(middleware.CORS([]string{"*"}))

	// 5. Health & Observability Endpoints
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{
			"service": constants.ServiceAPIGateway,
			"version": buildVersion,
			"commit":  buildCommit,
			"built":   buildTime,
		}, "Version Info")
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, nil, "API Gateway is healthy")
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Verify DB & Redis connections for readiness
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var dbStatus, redisStatus string = "ok", "ok"
		ready := true

		if db != nil {
			if err := database.PingPostgres(ctx, db); err != nil {
				dbStatus = "down"
				ready = false
			}
		} else {
			dbStatus = "not_configured"
		}

		if rdb != nil {
			if err := database.PingRedis(ctx, rdb); err != nil {
				redisStatus = "down"
				ready = false
			}
		} else {
			redisStatus = "not_configured"
		}

		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}

		utils.WriteJSON(w, status, map[string]string{
			"postgres": dbStatus,
			"redis":    redisStatus,
		}, "Readiness Probe")
	})

	r.Get("/live", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, nil, "API Gateway is live")
	})

	if cfg.Metrics.Enabled {
		r.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	// OpenAPI Preparation
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for Swagger UI
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Swagger UI (Coming Soon)</h1></body></html>"))
	})

	// 6. Start HTTP Server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 7. Graceful Shutdown
	go func() {
		log.Info("Server listening", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.DrainTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("API Gateway exited gracefully")
}
