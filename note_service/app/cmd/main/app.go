package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"note_service/internal/config"
	handlers "note_service/internal/handlers/notes"
	taghandlers "note_service/internal/handlers/tags"
	noteservice "note_service/internal/service/notes"
	tagservice "note_service/internal/service/tags"
	"note_service/internal/storage/db"
	"note_service/pkg/docs"
	"note_service/pkg/logging"
	"note_service/pkg/shutdown"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Info("note_service starting")
	cfg := config.GetConfig()

	storage, err := db.NewStorage(cfg)
	if err != nil {
		logger.Fatal("failed to initialize storage", "error", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"note_service"}`))
	})
	docs.Register(mux)

	noteHandler := handlers.Handler{
		Logger:      logger,
		NoteService: noteservice.NewService(storage),
	}
	noteHandler.Register(mux)

	tagHandler := taghandlers.Handler{
		Logger:     logger,
		TagService: tagservice.NewService(storage),
	}
	tagHandler.Register(mux)

	start(mux, logger, cfg)
}

func start(handler http.Handler, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal("failed to resolve app dir", "error", err)
		}

		socketPath := filepath.Join(appDir, "app.sock")
		logger.Info("listen unix socket", "socket_path", socketPath)
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal("failed to listen on socket", "error", err, "socket_path", socketPath)
		}
	} else {
		var err error
		logger.Info("listen tcp", "bind_ip", cfg.Listen.BindIP, "port", cfg.Listen.Port)
		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		if err != nil {
			logger.Fatal("failed to listen on tcp", "error", err, "bind_ip", cfg.Listen.BindIP, "port", cfg.Listen.Port)
		}
	}

	server = &http.Server{
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Gracful(
		[]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server,
	)

	logger.Info("server started")
	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal("server failed", "error", err)
		}
	}
}
