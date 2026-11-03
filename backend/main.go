package main

import (
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/a-float/mowik/handlers"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// Public routes
		e.Router.HandleFunc("/api/translate", handlers.TranslateHandler(e.App)).Methods(http.MethodPost)
		e.Router.HandleFunc("/api/enrich", handlers.EnrichHandler(e.App)).Methods(http.MethodPost)

		// Protected routes (require auth)
		// Lessons
		e.Router.HandleFunc("/api/lessons", handlers.LessonsHandler(e.App)).Methods(http.MethodGet, http.MethodPost)
		e.Router.HandleFunc("/api/lessons/{id}", handlers.LessonByIDHandler(e.App)).Methods(http.MethodGet, http.MethodDelete)

		// Favorites
		e.Router.HandleFunc("/api/favorites", handlers.FavoritesHandler(e.App)).Methods(http.MethodGet, http.MethodPost)
		e.Router.HandleFunc("/api/favorites/{id}", handlers.FavoriteByIDHandler(e.App)).Methods(http.MethodDelete)

		// Progress
		e.Router.HandleFunc("/api/progress", handlers.ProgressHandler(e.App)).Methods(http.MethodGet, http.MethodPost)
		e.Router.HandleFunc("/api/progress/vocab/{vocab_id}", handlers.ProgressByVocabHandler(e.App)).Methods(http.MethodGet, http.MethodPut, http.MethodPatch)
		e.Router.HandleFunc("/api/progress/review", handlers.GetWordsForReview(e.App)).Methods(http.MethodGet)
		e.Router.HandleFunc("/api/progress/stats", handlers.GetProgressStats(e.App)).Methods(http.MethodGet)

		// Quiz
		e.Router.HandleFunc("/api/quiz", handlers.QuizHandler(e.App)).Methods(http.MethodPost)
		e.Router.HandleFunc("/api/quiz/submit", handlers.QuizHandler(e.App)).Methods(http.MethodPost)

		return nil
	})

	log.Println("Mówik backend starting on :8080")
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}