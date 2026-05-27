package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"support_chat/internal/transport"
	"support_chat/internal/transport/ws"
	"support_chat/pkg/closer"
	"support_chat/pkg/config"
	"support_chat/pkg/logger"
	"syscall"
	"time"
)

type App struct {
	HttpPort string
	closer   *closer.Closer
	log      *slog.Logger
	server   *http.Server
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		return nil, fmt.Errorf("app.New load config error: %w", err)
	}

	logger.Setup(cfg.AppEnv)

	logs := logger.With("service", "support_chat")
	logger.WithContext(ctx, logs)
	logs.Info("initializing layers", "env", cfg.AppEnv, "port", cfg.HTTPPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/support", transport.LoggingMiddleware(logs, ws.HandleConnections))

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	cl := closer.New()

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing http server")
		return httpServer.Shutdown(ctx)
	})

	return &App{
		HttpPort: cfg.HTTPPort,
		log:      logs,
		server:   httpServer,
		closer:   cl,
	}, nil
}

func (a *App) Run() {
	errCh := make(chan error)

	go func() {
		a.log.Info("starting http server")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("app.Run: %w", err)
		}
	}()

	a.log.Info("App.Run starting server", "port", a.HttpPort)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		a.log.Error("app.run server startup failed", "error", err)
	case sig := <-quit:
		a.log.Info("app.run server shutdown", "signal", sig)
	}

	a.log.Info("shutting down servers")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.closer.Close(shutdownCtx); err != nil {
		a.log.Error("app.Run shutdown failed", "error", err)
	}

	fmt.Println("Server Stopped")
}
