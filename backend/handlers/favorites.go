package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// FavoriteRequest represents a favorite creation request.
type FavoriteRequest struct {
	VocabularyID string `json:"vocabulary_id"`
}

// FavoriteResponse represents a favorite with vocabulary details.
type FavoriteResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	VocabularyID string `json:"vocabulary_id"`
	Word         string `json:"word"`
	Translation  string `json:"translation"`
	Language     string `json:"language"`
	PartOfSpeech string `json:"part_of_speech"`
	Difficulty   string `json:"difficulty"`
	Example      string `json:"example"`
	CreatedAt    string `json:"created_at"`
}

// FavoritesHandler handles favorites CRUD operations.
func FavoritesHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		switch r.Method {
		case http.MethodGet:
			listFavorites(app, w, r, userID)
		case http.MethodPost:
			addFavorite(app, w, r, userID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// FavoriteByIDHandler handles individual favorite operations.
func FavoriteByIDHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		id := strings.TrimPrefix(r.URL.Path, "/api/favorites/")
		if id == "" {
			http.Error(w, "Favorite ID required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			deleteFavorite(app, w, r, userID, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listFavorites(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID string) {
	page := 1
	limit := 50
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}

	filter := fmt.Sprintf("user_id = '%s'", userID)
	records, err := app.Dao().FindRecordsByExpr("favorites", filter, "-created_at", limit, (page-1)*limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list favorites: %v", err), http.StatusInternalServerError)
		return
	}

	var favorites []FavoriteResponse
	for _, rec := range records {
		vocabID := rec.GetString("vocabulary_id")
		vocab, _ := app.Dao().FindRecordById("vocabulary", vocabID)
		fav := FavoriteResponse{
			ID:           rec.Id,
			UserID:       rec.GetString("user_id"),
			VocabularyID: vocabID,
			CreatedAt:    rec.GetString("created_at"),
		}
		if vocab != nil {
			fav.Word = vocab.GetString("word")
			fav.Translation = vocab.GetString("translation")
			fav.Language = vocab.GetString("language")
			fav.PartOfSpeech = vocab.GetString("part_of_speech")
			fav.Difficulty = vocab.GetString("difficulty")
			fav.Example = vocab.GetString("example_sentence")
		}
		favorites = append(favorites, fav)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func addFavorite(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID string) {
	var req FavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.VocabularyID == "" {
		http.Error(w, "vocabulary_id is required", http.StatusBadRequest)
		return
	}

	// Check if vocabulary exists
	vocab, err := app.Dao().FindRecordById("vocabulary", req.VocabularyID)
	if err != nil {
		http.Error(w, "Vocabulary not found", http.StatusNotFound)
		return
	}

	// Check if already favorited
	existing, err := app.Dao().FindFirstRecordByExpr("favorites", fmt.Sprintf("user_id = '%s' && vocabulary_id = '%s'", userID, req.VocabularyID))
	if err == nil && existing != nil {
		http.Error(w, "Already in favorites", http.StatusConflict)
		return
	}

	// Create favorite
	record := models.NewRecord()
	record.Set("collection", "favorites")
	record.Set("user_id", userID)
	record.Set("vocabulary_id", req.VocabularyID)

	if err := app.Dao().SaveRecord(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add favorite: %v", err), http.StatusInternalServerError)
		return
	}

	resp := FavoriteResponse{
		ID:           record.Id,
		UserID:       userID,
		VocabularyID: req.VocabularyID,
		Word:         vocab.GetString("word"),
		Translation:  vocab.GetString("translation"),
		Language:     vocab.GetString("language"),
		PartOfSpeech: vocab.GetString("part_of_speech"),
		Difficulty:   vocab.GetString("difficulty"),
		Example:      vocab.GetString("example_sentence"),
		CreatedAt:    record.GetString("created_at"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func deleteFavorite(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID, id string) {
	record, err := app.Dao().FindRecordById("favorites", id)
	if err != nil {
		http.Error(w, "Favorite not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if record.GetString("user_id") != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := app.Dao().DeleteRecord(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete favorite: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}