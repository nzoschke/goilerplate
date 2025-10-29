package main

import (
	"log/slog"
	"net/http"

	"github.com/templui/goilerplate/internal/app"
	"github.com/templui/goilerplate/internal/config"
	"github.com/templui/goilerplate/internal/logger"
	"github.com/templui/goilerplate/internal/routes"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.IsDevelopment(), cfg.SentryDSN)

	app, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		panic(err)
	}
	defer func() {
		closeErr := app.Close()
		if closeErr != nil {
			slog.Error("failed to close app", "error", closeErr)
		}
	}()

	handler := routes.SetupRoutes(app)
	slog.Info("server starting", "port", cfg.Port, "env", cfg.AppEnv, "url", "http://localhost:"+cfg.Port)

	err = http.ListenAndServe(":"+cfg.Port, handler)
	if err != nil {
		slog.Error("server failed", "error", err)
		panic(err)
	}
}
