package main

import (
	"errors"
	"fmt"
	"myproject/app/internal/handlers/auth"
	"myproject/app/internal/handlers/categories"

	"myproject/app/pkg/handlers/metric"
	"myproject/app/pkg/logging"

	"github.com/julienschmidt/httprouter"

	"myproject/app/internal/config"
	"myproject/app/pkg/cache/freecache"
	"myproject/app/pkg/shutdown"

	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()
	logger.Println("Logger инициализирован")

	logger.Println("config инициализирован")
	cfg := config.GetConfig()

	logger.Println("router инициализирован")
	router := httprouter.New()

	logger.Println("cache инициализирован")
	refreshTokenCache := freecache.NewCacheRepo(104857600)

	logger.Println("create and regiser handlers")
	authHandler := auth.Handler{RTCache: refreshTokenCache, Logger: logger}
	authHandler.Register(router)

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	categoriesHandler := categories.Handler{Logger: logger}
	categoriesHandler.Register(router)

	logger.Println("start application")
	start(router, logger, cfg)
}

func start(router *httprouter.Router, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}

		socketPath := filepath.Join(appDir, "app.sock")
		logger.Infof("socket path: %s", socketPath)

		logger.Info("create and listen unix socket")
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Infof("bind application to host: %s and port: %s", cfg.Listen.BindIP, cfg.Listen.Port)

		var err error

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		if err != nil {
			logger.Fatal(err)
		}
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Gracful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM}, server)

	logger.Println("application initialized and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}
