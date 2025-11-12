package main

import (
	"fmt"
	"net/http"

	"github.com/PinceredCoder/RestGo/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	taskHandler := handlers.NewTaskHandler()

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
