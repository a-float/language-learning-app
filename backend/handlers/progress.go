package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// ProgressRequest represents a progress update request.
type ProgressRequest struct {
	VocabularyID string `json:"vocabulary_id"`
	KnowledgeLevel int  `json:"knowledge_level"` // 0-5
}

// ProgressResponse represents user vocabulary progress.
type ProgressResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	VocabularyID   string `json:"vocabulary_id"`
	KnowledgeLevel int    `json:"knowledge_level"`
	LastReviewed   string `json:"last_reviewed"`
	ReviewCount    int    `json:"review_count"`
	Word           string `json:"word"`
	Translation    string `json:"translation"`
	Language       string `json:"language"`
}

// ProgressHandler handles user vocabulary progress operations.
func ProgressHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Require authentication
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		switch r.Method {
		case http.MethodGet:
			listProgress(app, w, r, userID)
		case http.MethodPost:
			updateProgress(app, w, r, userID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// ProgressByVocabHandler handles progress for a specific vocabulary item.
func ProgressByVocabHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		vocabID := strings.TrimPrefix(r.URL.Path, "/api/progress/vocab/")
		if vocabID == "" {
			http.Error(w, "Vocabulary ID required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getProgress(app, w, r, userID, vocabID)
		case http.MethodPut, http.MethodPatch:
			updateProgressSpecific(app, w, r, userID, vocabID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listProgress(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID string) {
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

	// Filter by knowledge level if provided
	filter := fmt.Sprintf("user_id = '%s'", userID)
	if level := r.URL.Query().Get("level"); level != "" {
		filter += fmt.Sprintf(" && knowledge_level = %s", level)
	}

	records, err := app.Dao().FindRecordsByExpr("user_vocabulary_progress", filter, "-last_reviewed", limit, (page-1)*limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list progress: %v", err), http.StatusInternalServerError)
		return
	}

	var progress []ProgressResponse
	for _, rec := range records {
		vocabID := rec.GetString("vocabulary_id")
		vocab, _ := app.Dao().FindRecordById("vocabulary", vocabID)
		word, translation, lang := "", "", ""
		if vocab != nil {
			word = vocab.GetString("word")
			translation = vocab.GetString("translation")
			lang = vocab.GetString("language")
		}

		progress = append(progress, ProgressResponse{
			ID:             rec.Id,
			UserID:         rec.GetString("user_id"),
			VocabularyID:   vocabID,
			KnowledgeLevel: int(rec.GetInt("knowledge_level")),
			LastReviewed:   rec.GetString("last_reviewed"),
			ReviewCount:    int(rec.GetInt("review_count")),
			Word:           word,
			Translation:    translation,
			Language:       lang,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

func getProgress(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID, vocabID string) {
	record, err := app.Dao().FindFirstRecordByExpr("user_vocabulary_progress", fmt.Sprintf("user_id = '%s' && vocabulary_id = '%s'", userID, vocabID))
	if err != nil {
		// Return default progress (not started)
		vocab, _ := app.Dao().FindRecordById("vocabulary", vocabID)
		resp := ProgressResponse{
			UserID:         userID,
			VocabularyID:   vocabID,
			KnowledgeLevel: 0,
			ReviewCount:    0,
		}
		if vocab != nil {
			resp.Word = vocab.GetString("word")
			resp.Translation = vocab.GetString("translation")
			resp.Language = vocab.GetString("language")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	vocab, _ := app.Dao().FindRecordById("vocabulary", vocabID)
	resp := ProgressResponse{
		ID:             record.Id,
		UserID:         record.GetString("user_id"),
		VocabularyID:   record.GetString("vocabulary_id"),
		KnowledgeLevel: int(record.GetInt("knowledge_level")),
		LastReviewed:   record.GetString("last_reviewed"),
		ReviewCount:    int(record.GetInt("review_count")),
	}
	if vocab != nil {
		resp.Word = vocab.GetString("word")
		resp.Translation = vocab.GetString("translation")
		resp.Language = vocab.GetString("language")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func updateProgress(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID string) {
	var req ProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.VocabularyID == "" {
		http.Error(w, "vocabulary_id is required", http.StatusBadRequest)
		return
	}

	if req.KnowledgeLevel < 0 || req.KnowledgeLevel > 5 {
		http.Error(w, "knowledge_level must be between 0 and 5", http.StatusBadRequest)
		return
	}

	updateProgressRecord(app, w, userID, req.VocabularyID, req.KnowledgeLevel)
}

func updateProgressSpecific(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID, vocabID string) {
	var req ProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.KnowledgeLevel < 0 || req.KnowledgeLevel > 5 {
		http.Error(w, "knowledge_level must be between 0 and 5", http.StatusBadRequest)
		return
	}

	updateProgressRecord(app, w, userID, vocabID, req.KnowledgeLevel)
}

func updateProgressRecord(app *pocketbase.PocketBase, w http.ResponseWriter, userID, vocabID string, level int) {
	record, err := app.Dao().FindFirstRecordByExpr("user_vocabulary_progress", fmt.Sprintf("user_id = '%s' && vocabulary_id = '%s'", userID, vocabID))
	now := time.Now().UTC().Format(time.RFC3339)

	if err != nil || record == nil {
		// Create new progress record
		record = models.NewRecord()
		record.Set("collection", "user_vocabulary_progress")
		record.Set("user_id", userID)
		record.Set("vocabulary_id", vocabID)
		record.Set("knowledge_level", level)
		record.Set("last_reviewed", now)
		record.Set("review_count", 1)
	} else {
		// Update existing
		record.Set("knowledge_level", level)
		record.Set("last_reviewed", now)
		record.Set("review_count", record.GetInt("review_count")+1)
	}

	if err := app.Dao().SaveRecord(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save progress: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated progress
	vocab, _ := app.Dao().FindRecordById("vocabulary", vocabID)
	resp := ProgressResponse{
		ID:             record.Id,
		UserID:         userID,
		VocabularyID:   vocabID,
		KnowledgeLevel: level,
		LastReviewed:   now,
		ReviewCount:    int(record.GetInt("review_count")),
	}
	if vocab != nil {
		resp.Word = vocab.GetString("word")
		resp.Translation = vocab.GetString("translation")
		resp.Language = vocab.GetString("language")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetWordsForReview returns words due for review based on spaced repetition.
func GetWordsForReview(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
				limit = v
			}
		}

		// Simple spaced repetition: review words with lower knowledge level more frequently
		// Level 0-1: review daily, Level 2: every 3 days, Level 3: weekly, Level 4: biweekly, Level 5: monthly
		// For simplicity, just return words with lowest knowledge level that haven't been reviewed recently
		filter := fmt.Sprintf("user_id = '%s' && knowledge_level < 5", userID)
		records, err := app.Dao().FindRecordsByExpr("user_vocabulary_progress", filter, "knowledge_level,last_reviewed", limit, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get review words: %v", err), http.StatusInternalServerError)
			return
		}

		var progress []ProgressResponse
		for _, rec := range records {
			vocabID := rec.GetString("vocabulary_id")
			vocab, _ := app.Dao().FindRecordById("vocabulary", vocabID)
			word, translation, lang := "", "", ""
			if vocab != nil {
				word = vocab.GetString("word")
				translation = vocab.GetString("translation")
				lang = vocab.GetString("language")
			}

			progress = append(progress, ProgressResponse{
				ID:             rec.Id,
				UserID:         rec.GetString("user_id"),
				VocabularyID:   vocabID,
				KnowledgeLevel: int(rec.GetInt("knowledge_level")),
				LastReviewed:   rec.GetString("last_reviewed"),
				ReviewCount:    int(rec.GetInt("review_count")),
				Word:           word,
				Translation:    translation,
				Language:       lang,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(progress)
	}
}

// GetProgressStats returns statistics about user's vocabulary progress.
func GetProgressStats(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		records, err := app.Dao().FindRecordsByExpr("user_vocabulary_progress", fmt.Sprintf("user_id = '%s'", userID))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
			return
		}

		stats := map[string]int{
			"total_words":    len(records),
			"level_0_new":    0,
			"level_1_learning": 0,
			"level_2_familiar": 0,
			"level_3_known":  0,
			"level_4_mastered": 0,
			"level_5_fluent": 0,
			"total_reviews":  0,
		}

		for _, rec := range records {
			level := int(rec.GetInt("knowledge_level"))
			stats["total_reviews"] += int(rec.GetInt("review_count"))
			switch level {
			case 0:
				stats["level_0_new"]++
			case 1:
				stats["level_1_learning"]++
			case 2:
				stats["level_2_familiar"]++
			case 3:
				stats["level_3_known"]++
			case 4:
				stats["level_4_mastered"]++
			case 5:
				stats["level_5_fluent"]++
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}