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

	"github.com/julienschmidt/httprouter"

	"file_service/internal/config"
	filehandlers "file_service/internal/handlers/files"
	fileservice "file_service/internal/service/files"
	miniostorage "file_service/internal/storage/minio"
	"file_service/pkg/logging"
	"file_service/pkg/shutdown"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Info("file_service starting")

	cfg := config.GetConfig()

	storage, err := miniostorage.NewStorage(
		cfg.Minio.Endpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.UseSSL,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to initialize storage", "error", err)
	}

	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"file_service"}`))
	})

	fileHandler := filehandlers.Handler{
		Logger:             logger,
		MaxUploadSizeBytes: cfg.Upload.MaxFileSizeMB * 1024 * 1024,
		FileService: fileservice.NewService(
			storage,
			logger,
			cfg.Upload.MaxFileSizeMB*1024*1024,
		),
	}
	fileHandler.Register(router)

	start(router, logger, cfg)
}

func start(router *httprouter.Router, logger logging.Logger, cfg *config.Config) {
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
		Handler:      router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
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
