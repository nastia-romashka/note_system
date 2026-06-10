package main

import (
	"errors"
	"fmt"
	categoryclient "myproject/internal/client/category"
	fileclient "myproject/internal/client/file"
	noteclient "myproject/internal/client/note"
	searchclient "myproject/internal/client/search"
	userclient "myproject/internal/client/user"
	"myproject/internal/handlers/actionlog"
	"myproject/internal/handlers/auth"
	"myproject/internal/handlers/categories"
	"myproject/internal/handlers/files"
	"myproject/internal/handlers/notes"
	"myproject/internal/handlers/passthrough"
	"myproject/internal/handlers/profile"
	searchhandler "myproject/internal/handlers/search"
	"myproject/internal/handlers/tags"

	"myproject/internal/config"
	"myproject/pkg/docs"
	"myproject/pkg/handlers/metric"
	"myproject/pkg/logging"
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
	logger.Println("logger initialized")

	logger.Println("config initialized")
	cfg := config.GetConfig()

	logger.Println("router initialized")
	router := http.NewServeMux()
	router.HandleFunc(http.MethodGet+" /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"api_service"}`))
	})
	docs.Register(router)

	logger.Println("create and register handlers")
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
	searchService := searchclient.NewService(
		cfg.SearchService.URL,
		logger,
	)
	passthroughHandler := passthrough.NewHandler(
		logger,
		cfg.CategoryService.URL,
		cfg.NoteService.URL,
		cfg.UserService.URL,
		cfg.SearchService.URL,
	)
	passthroughHandler.Register(router)
	profileHandler := profile.Handler{
		Logger:          logger,
		UserService:     userService,
		CategoryService: categoryService,
	}

	noteService := noteclient.NewService(
		cfg.NoteService.URL,
		"notes",
		logger,
	)
	fileService := fileclient.NewService(
		cfg.FileService.URL,
		"files",
		logger,
	)

	profileHandler.NoteService = noteService
	categoriesHandler := categories.Handler{
		Logger:          logger,
		CategoryService: categoryService,
		FileService:     fileService,
		NoteService:     noteService,
		SearchService:   searchService,
		ActionRecorder:  actionRecorder,
	}
	categoriesHandler.Register(router)

	notesHandler := notes.Handler{
		Logger:          logger,
		CategoryService: categoryService,
		FileService:     fileService,
		NoteService:     noteService,
		SearchService:   searchService,
		ActionRecorder:  actionRecorder,
	}
	notesHandler.Register(router)

	tagsHandler := tags.Handler{
		Logger:         logger,
		NoteService:    noteService,
		ActionRecorder: actionRecorder,
	}
	tagsHandler.Register(router)

	profileHandler.FileService = fileService
	profileHandler.Register(router)

	filesHandler := files.Handler{
		Logger:         logger,
		NoteService:    noteService,
		FileService:    fileService,
		ActionRecorder: actionRecorder,
	}
	filesHandler.Register(router)

	searchNotesHandler := searchhandler.Handler{
		Logger:          logger,
		SearchService:   searchService,
		CategoryService: categoryService,
		NoteService:     noteService,
	}
	searchNotesHandler.Register(router)

	logger.Println("start application")
	start(router, logger, cfg)
}

func start(handler http.Handler, logger logging.Logger, cfg *config.Config) {
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
		Handler:      handler,
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
