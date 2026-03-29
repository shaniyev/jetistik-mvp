package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/auth"
	"jetistik/internal/platform/config"
	"jetistik/internal/platform/db"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
	"jetistik/internal/user"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	if cfg.IsDev() {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	slog.SetDefault(logger)

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	slog.Info("connected to database")

	// Wire modules
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authHandler := auth.NewHandler(authSvc, cfg.JWTRefreshTTL, !cfg.IsDev())

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.PublicBaseURL))

	r.Route("/api/v1", func(r chi.Router) {
		// Health
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			err := pool.Ping(r.Context())
			if err != nil {
				response.Error(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "database is not reachable")
				return
			}
			response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		// Public auth routes (rate-limited)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimit(10, time.Minute))
			r.Mount("/auth", authHandler.Routes())
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))

			r.Mount("/profile", userHandler.ProfileRoutes())
			r.Mount("/teacher/students", userHandler.TeacherStudentRoutes())
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}
