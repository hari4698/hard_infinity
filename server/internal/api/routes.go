package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hari4698/hardinfinity/internal/auth"
	"github.com/hari4698/hardinfinity/internal/handlers"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	//Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Should be restricted in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//Public Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	//API routes with authentication
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.Middleware)

		//Challenges
		r.Route("/challenges", func(r chi.Router) {
			r.Get("/", handlers.GetChallenges)
			r.Post("/", handlers.CreateChallenge)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handlers.GetChallenge)
				r.Put("/", handlers.UpdateChallenge)
				r.Delete("/", handlers.DeleteChallenge)
				r.Post("/reset", handlers.ResetChallenge)
				r.Get("/progress", handlers.GetChallengeProgress)
			})
		})

		//Sections
		r.Route("/challenges/{challengeId}/sections", func(r chi.Router) {
			r.Get("/", handlers.GetSections)
			r.Post("/", handlers.CreateSection)
		})

		r.Route("/sections/{id}", func(r chi.Router) {
			r.Put("/", handlers.UpdateSection)
			r.Delete("/", handlers.DeleteSection)
			r.Put("/order", handlers.ReorderSection)
		})

		// Tasks
		r.Route("/sections/{sectionId}/tasks", func(r chi.Router) {
			r.Get("/", handlers.GetTasks)
			r.Post("/", handlers.CreateTask)
		})

		r.Route("/tasks/{id}", func(r chi.Router) {
			r.Put("/", handlers.UpdateTask)
			r.Delete("/", handlers.DeleteTask)
			r.Put("/order", handlers.ReorderTask)
		})
		// Daily Entries
		r.Route("/challenges/{challengeId}/entries", func(r chi.Router) {
			r.Get("/", handlers.GetDailyEntries)
			r.Post("/", handlers.CreateOrUpdateTodayEntry)

			r.Route("/{day}", func(r chi.Router) {
				r.Get("/", handlers.GetDailyEntry)
				r.Put("/", handlers.UpdateDailyEntry)
			})
		})

		// Measurements
		r.Route("/challenges/{challengeId}/measurements", func(r chi.Router) {
			r.Get("/", handlers.GetMeasurements)
			r.Post("/", handlers.AddMeasurement)
		})

		r.Route("/measurements/{id}", func(r chi.Router) {
			r.Put("/", handlers.UpdateMeasurement)
		})

	})

	return r
}
