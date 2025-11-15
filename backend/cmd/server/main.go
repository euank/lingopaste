package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/lingopaste/backend/internal/cache"
	"github.com/lingopaste/backend/internal/config"
	"github.com/lingopaste/backend/internal/db"
	"github.com/lingopaste/backend/internal/handlers"
	"github.com/lingopaste/backend/internal/middleware"
	"github.com/lingopaste/backend/internal/storage"
	"github.com/lingopaste/backend/internal/translate"
)

type Server struct {
	cfg          *config.Config
	db           *db.DynamoDB
	storage      *storage.S3Storage
	cache        *cache.LRUCache
	translator   *translate.OpenAITranslator
	pasteHandler *handlers.PasteHandler
	router       *mux.Router
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dynamoDB, err := db.NewDynamoDB(
		ctx,
		cfg.AWSRegion,
		cfg.DynamoDBAccountsTable,
		cfg.DynamoDBPastesTable,
		cfg.DynamoDBRateLimitsTable,
	)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}

	s3Storage, err := storage.NewS3Storage(ctx, cfg.AWSRegion, cfg.S3BucketName)
	if err != nil {
		log.Fatalf("Failed to initialize S3: %v", err)
	}

	lruCache := cache.NewLRUCache(cfg.CacheSize)
	translator := translate.NewOpenAITranslator(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	pasteHandler := handlers.NewPasteHandler(dynamoDB, s3Storage, lruCache, translator, cfg.MaxPasteLength)

	server := &Server{
		cfg:          cfg,
		db:           dynamoDB,
		storage:      s3Storage,
		cache:        lruCache,
		translator:   translator,
		pasteHandler: pasteHandler,
		router:       mux.NewRouter(),
	}

	server.setupRoutes()

	corsMiddleware := middleware.NewCORS(cfg.FrontendURL)
	rateLimiter := middleware.NewRateLimiter(dynamoDB)

	handler := corsMiddleware.Handler(
		middleware.Logger(
			middleware.ExtractIP(
				rateLimiter.Middleware(server.router),
			),
		),
	)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/pastes", s.pasteHandler.Create).Methods("POST")
	api.HandleFunc("/pastes/{id}", s.pasteHandler.Get).Methods("GET")
	api.HandleFunc("/pastes/{id}/translate", s.pasteHandler.Translate).Methods("GET")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
