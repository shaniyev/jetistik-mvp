package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"jetistik/internal/admin"
	"jetistik/internal/audit"
	"jetistik/internal/auth"
	"jetistik/internal/batch"
	"jetistik/internal/certificate"
	"jetistik/internal/event"
	"jetistik/internal/organization"
	"jetistik/internal/platform/config"
	"jetistik/internal/platform/db"
	"jetistik/internal/platform/middleware"
	"jetistik/internal/platform/response"
	"jetistik/internal/storage"
	tmpl "jetistik/internal/template"
	"jetistik/internal/user"
	"jetistik/internal/worker"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	workerMode := flag.Bool("worker", false, "run as async worker instead of HTTP server")
	flag.Parse()

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

	// MinIO storage
	storageClient, err := storage.NewClient(
		cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey,
		cfg.MinioBucket, cfg.MinioUseSSL,
	)
	if err != nil {
		return fmt.Errorf("connect to minio: %w", err)
	}
	slog.Info("connected to MinIO", "bucket", cfg.MinioBucket)

	// Worker mode: run Asynq worker server
	if *workerMode {
		srv, err := worker.NewServer(pool, storageClient, cfg)
		if err != nil {
			return fmt.Errorf("create worker: %w", err)
		}

		go func() {
			<-ctx.Done()
			slog.Info("shutting down worker")
			srv.Shutdown()
		}()

		return srv.Run()
	}

	// HTTP server mode

	// Asynq client for enqueueing tasks
	asynqClient, err := worker.NewClient(cfg.RedisURL)
	if err != nil {
		slog.Warn("failed to create asynq client, generation will be unavailable", "error", err)
	} else {
		defer asynqClient.Close()
	}

	// SSE handler for progress streaming
	sseHandler, err := worker.NewSSEHandler(cfg.RedisURL)
	if err != nil {
		slog.Warn("failed to create SSE handler", "error", err)
	} else {
		defer sseHandler.Close()
		sseHandler.LogInfo()
	}

	// Wire modules
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authHandler := auth.NewHandler(authSvc, cfg.JWTRefreshTTL, !cfg.IsDev())

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	auditRepo := audit.NewRepository(pool)
	auditSvc := audit.NewService(auditRepo)
	auditHandler := audit.NewHandler(auditSvc)

	orgRepo := organization.NewRepository(pool)
	orgSvc := organization.NewService(orgRepo)
	orgHandler := organization.NewHandler(orgSvc)

	eventRepo := event.NewRepository(pool)
	eventSvc := event.NewService(eventRepo)
	eventHandler := event.NewHandler(eventSvc, orgSvc, auditSvc)

	tmplRepo := tmpl.NewRepository(pool)
	tmplSvc := tmpl.NewService(tmplRepo, storageClient)
	tmplHandler := tmpl.NewHandler(tmplSvc, auditSvc)

	batchRepo := batch.NewRepository(pool)
	batchSvc := batch.NewService(batchRepo, storageClient)
	var enqueuer batch.TaskEnqueuer
	if asynqClient != nil {
		enqueuer = asynqClient
	}
	batchHandler := batch.NewHandler(batchSvc, tmplSvc, auditSvc, enqueuer)

	certRepo := certificate.NewRepository(pool)
	certSvc := certificate.NewService(certRepo, storageClient, cfg.PublicBaseURL)
	certHandler := certificate.NewHandler(certSvc, auditSvc)

	adminSvc := admin.NewService(pool)
	adminHandler := admin.NewHandler(adminSvc)

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

		// Public routes (rate-limited)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimit(10, time.Minute))
			r.Mount("/", certHandler.PublicRoutes())
			r.Mount("/p", userHandler.PublicProfileRoutes())
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTSecret))

			r.Mount("/profile", userHandler.ProfileRoutes())
			r.Mount("/teacher/students", userHandler.TeacherStudentRoutes())
			r.Mount("/teacher/certificates", userHandler.TeacherCertificateRoutes())

			// Staff routes
			r.Route("/staff", func(r chi.Router) {
				r.Use(middleware.RequireRole("staff", "admin"))

				r.Mount("/events", eventHandler.StaffRoutes())

				// Template upload/delete on events
				r.Post("/events/{id}/template", tmplHandler.Upload)
				r.Delete("/events/{id}/template", tmplHandler.Delete)
				r.Get("/events/{id}/template", tmplHandler.GetByEvent)

				// Batch upload on events
				r.Get("/events/{id}/batches", batchHandler.ListByEvent)
			r.Post("/events/{id}/batches", batchHandler.Upload)

				// Batch operations
				r.Get("/batches/{id}", batchHandler.GetByID)
				r.Patch("/batches/{id}/mapping", batchHandler.UpdateMapping)
				r.Post("/batches/{id}/generate", batchHandler.Generate)
				r.Delete("/batches/{id}", batchHandler.Delete)

				// SSE progress endpoint
				if sseHandler != nil {
					r.Get("/batches/{id}/progress", sseHandler.ServeProgress)
				}

				// Certificates per event
				r.Route("/events/{id}/certificates", func(r chi.Router) {
					r.Mount("/", certHandler.StaffCertificateRoutes())
				})

				// Individual certificate operations
				r.Mount("/certificates", certHandler.StaffCertificateItemRoutes())

				// Audit log
				r.Mount("/audit-log", auditHandler.Routes())
			})

			// Admin routes
			r.Route("/admin", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))

				r.Mount("/organizations", orgHandler.AdminRoutes())
				r.Mount("/audit-log", auditHandler.Routes())
				r.Mount("/", adminHandler.Routes())
			})
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
