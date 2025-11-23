package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/PinceredCoder/restGo/internal/database"
	"github.com/PinceredCoder/restGo/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lmittmann/tint"
)

func main() {
	// Set up beautiful colored logging with tint
	logger := slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen, // "3:04PM" format
		}),
	)
	slog.SetDefault(logger)

	logger.Info("Starting restGo API server")

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	logger.Info("Connecting to MongoDB", "uri", "mongodb://127.0.0.1:27017", "database", "tasks")
	db, err := database.NewMongoDatabase(context.Background(), "mongodb://127.0.0.1:27017", "tasks")
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err)
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Disconnect(context.Background())
	logger.Info("Successfully connected to MongoDB")

	taskHandler := handlers.NewTaskHandler(db, logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.GetAll)
			r.Post("/", taskHandler.Create)
			r.Get("/{id}", taskHandler.GetByID)
			r.Put("/{id}", taskHandler.Update)
			r.Delete("/{id}", taskHandler.Delete)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	port := ":8080"
	fmt.Printf("Server starting on %s\n", port)
	fmt.Println("API endpoints:")
	fmt.Println("  GET    /health")
	fmt.Println("  GET    /api/v1/tasks")
	fmt.Println("  POST   /api/v1/tasks")
	fmt.Println("  GET    /api/v1/tasks/{id}")
	fmt.Println("  PUT    /api/v1/tasks/{id}")
	fmt.Println("  DELETE /api/v1/tasks/{id}")

	if err := http.ListenAndServe(port, r); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
