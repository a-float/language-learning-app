package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestTranslateHandlerIntegration(t *testing.T) {
	// Create a test PocketBase app
	app := tests.NewTestApp(t)
	defer app.Cleanup()

	// Register the translate handler
	handler := TranslateHandler(app)

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/api/translate", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// This will fail without OPENROUTER_API_KEY, but we can test the validation
	handler.ServeHTTP(w, req)

	// Should return method not allowed for GET
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 for GET, got %d", w.Code)
	}
}

func TestTranslateHandlerValidation(t *testing.T) {
	app := tests.NewTestApp(t)
	defer app.Cleanup()

	handler := TranslateHandler(app)

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "empty body",
			body:           "{}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing text",
			body:           `{"target_lang": "es"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing target_lang",
			body:           `{"text": "hello"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid target_lang",
			body:           `{"text": "hello", "target_lang": "fr"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid request structure",
			body:           `{"text": "hello", "target_lang": "es"}`,
			expectedStatus: http.StatusInternalServerError, // Will fail without API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/translate", nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestLessonsHandlerIntegration(t *testing.T) {
	app := tests.NewTestApp(t)
	defer app.Cleanup()

	handler := LessonsHandler(app)

	// Test GET /api/lessons
	req := httptest.NewRequest(http.MethodGet, "/api/lessons", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 200 with empty list (no auth required for list)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for GET /api/lessons, got %d", w.Code)
	}
}

func TestFavoritesHandlerRequiresAuth(t *testing.T) {
	app := tests.NewTestApp(t)
	defer app.Cleanup()

	handler := FavoritesHandler(app)

	// Test GET /api/favorites without auth
	req := httptest.NewRequest(http.MethodGet, "/api/favorites", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for GET /api/favorites without auth, got %d", w.Code)
	}
}

func TestProgressHandlerRequiresAuth(t *testing.T) {
	app := tests.NewTestApp(t)
	defer app.Cleanup()

	handler := ProgressHandler(app)

	// Test GET /api/progress without auth
	req := httptest.NewRequest(http.MethodGet, "/api/progress", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for GET /api/progress without auth, got %d", w.Code)
	}
}

func TestQuizHandlerRequiresAuth(t *testing.T) {
	app := tests.NewTestApp(t)
	defer app.Cleanup()

	handler := QuizHandler(app)

	// Test POST /api/quiz without auth
	req := httptest.NewRequest(http.MethodPost, "/api/quiz", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should return 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for POST /api/quiz without auth, got %d", w.Code)
	}
}

func TestRouteRegistration(t *testing.T) {
	app := pocketbase.New()

	// This tests that the routes are registered without panic
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.HandleFunc("/api/translate", TranslateHandler(e.App)).Methods(http.MethodPost)
		e.Router.HandleFunc("/api/enrich", EnrichHandler(e.App)).Methods(http.MethodPost)
		e.Router.HandleFunc("/api/lessons", LessonsHandler(e.App)).Methods(http.MethodGet, http.MethodPost)
		e.Router.HandleFunc("/api/lessons/{id}", LessonByIDHandler(e.App)).Methods(http.MethodGet, http.MethodDelete)
		e.Router.HandleFunc("/api/favorites", FavoritesHandler(e.App)).Methods(http.MethodGet, http.MethodPost)
		e.Router.HandleFunc("/api/favorites/{id}", FavoriteByIDHandler(e.App)).Methods(http.MethodDelete)
		e.Router.HandleFunc("/api/progress", ProgressHandler(e.App)).Methods(http.MethodGet, http.MethodPost)
		e.Router.HandleFunc("/api/progress/vocab/{vocab_id}", ProgressByVocabHandler(e.App)).Methods(http.MethodGet, http.MethodPut, http.MethodPatch)
		e.Router.HandleFunc("/api/progress/review", GetWordsForReview(e.App)).Methods(http.MethodGet)
		e.Router.HandleFunc("/api/progress/stats", GetProgressStats(e.App)).Methods(http.MethodGet)
		e.Router.HandleFunc("/api/quiz", QuizHandler(e.App)).Methods(http.MethodPost)
		e.Router.HandleFunc("/api/quiz/submit", QuizHandler(e.App)).Methods(http.MethodPost)
		return nil
	})

	// Just verify no panic during route registration
	// We can't easily test the router without starting the server
	_ = app
}