package main

import (
	"errors"
	"fmt"
	categoryclient "myproject/internal/client/category"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	userclient "myproject/internal/client/user"
	"myproject/internal/handlers/actionlog"
	"myproject/internal/handlers/auth"
	"myproject/internal/handlers/categories"
	"myproject/internal/handlers/files"
	"myproject/internal/handlers/graph"
	"myproject/internal/handlers/notes"
	"myproject/internal/handlers/profile"
	"myproject/internal/handlers/tags"

	"myproject/pkg/handlers/metric"
	"myproject/pkg/logging"

	"github.com/julienschmidt/httprouter"

	"myproject/internal/config"
	"myproject/pkg/cache/freecache"
	"myproject/pkg/shutdown"

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
	router.HandlerFunc(http.MethodGet, "/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"api_service"}`))
	})

	logger.Println("cache инициализирован")
	refreshTokenCache := freecache.NewCacheRepo(104857600)

	logger.Println("create and regiser handlers")
	userService := userclient.NewService(
		cfg.UserService.URL,
		"users",
		logger,
	)
	actionRecorder := actionlog.Recorder{
		Logger:      logger,
		UserService: userService,
	}
	authHandler := auth.Handler{
		RTCache:     refreshTokenCache,
		Logger:      logger,
		UserService: userService,
	}
	authHandler.Register(router)

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	categoryService := categoryclient.NewService(
		cfg.CategoryService.URL,
		"categories",
		logger,
	)
	profileHandler := profile.Handler{
		Logger:          logger,
		UserService:     userService,
		CategoryService: categoryService,
	}

	categoriesHandler := categories.Handler{
		Logger:          logger,
		CategoryService: categoryService,
		ActionRecorder:  actionRecorder,
	}
	categoriesHandler.Register(router)
	graphHandler := graph.Handler{
		Logger:          logger,
		CategoryService: categoryService,
	}
	graphHandler.Register(router)

	noteService := noteclient.NewService(
		cfg.NoteService.URL,
		"notes",
		logger,
	)
	profileHandler.NoteService = noteService
	notesHandler := notes.Handler{
		Logger:          logger,
		CategoryService: categoryService,
		NoteService:     noteService,
		ActionRecorder:  actionRecorder,
	}
	notesHandler.Register(router)

	tagsHandler := tags.Handler{
		Logger:         logger,
		NoteService:    noteService,
		ActionRecorder: actionRecorder,
	}
	tagsHandler.Register(router)

	fileService := fileclient.NewService(
		cfg.FileService.URL,
		"files",
		logger,
	)
	profileHandler.FileService = fileService
	profileHandler.Register(router)

	filesHandler := files.Handler{
		Logger:         logger,
		NoteService:    noteService,
		FileService:    fileService,
		ActionRecorder: actionRecorder,
	}
	filesHandler.Register(router)

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
