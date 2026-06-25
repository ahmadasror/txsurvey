// Command server is the txsurvey entrypoint: structured logging, config load,
// migrations, connection pool, dependency wiring, and graceful shutdown.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/database"
	"github.com/ahmadasror/txsurvey/internal/handler"
	"github.com/ahmadasror/txsurvey/internal/logging"
	"github.com/ahmadasror/txsurvey/internal/repository"
	"github.com/ahmadasror/txsurvey/internal/router"
	"github.com/ahmadasror/txsurvey/internal/service"
	"github.com/ahmadasror/txsurvey/pkg/auth"
)

func main() {
	logging.Setup()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Dependency wiring (repos -> services -> handlers).
	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.SessionTTL)

	userRepo := repository.NewUserRepo(pool)
	formRepo := repository.NewFormRepo(pool)
	questionRepo := repository.NewQuestionRepo(pool)
	responseRepo := repository.NewResponseRepo(pool)

	authSvc := service.NewAuthService(cfg, userRepo)
	formSvc := service.NewFormService(formRepo, questionRepo)
	questionSvc := service.NewQuestionService(formRepo, questionRepo)
	responseSvc := service.NewResponseService(formRepo, questionRepo, responseRepo)

	h := &router.Handlers{
		Auth:     handler.NewAuthHandler(authSvc, jwtMgr, cfg),
		Form:     handler.NewFormHandler(formSvc),
		Question: handler.NewQuestionHandler(questionSvc),
		Public:   handler.NewPublicHandler(responseSvc),
	}

	r := router.Setup(cfg, h, jwtMgr)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", "port", cfg.ServerPort, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(timeoutCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
