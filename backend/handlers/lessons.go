package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// LessonRequest represents the incoming lesson creation request.
type LessonRequest struct {
	Title       string `json:"title"`
	SourceText  string `json:"source_text"`
	TargetLang  string `json:"target_lang"` // "es" or "it"
	Difficulty  string `json:"difficulty,omitempty"`
	SourceType  string `json:"source_type,omitempty"` // "text", "song", "article"
}

// LessonResponse represents a lesson with its vocabulary.
type LessonResponse struct {
	ID            string           `json:"id"`
	Title         string           `json:"title"`
	SourceText    string           `json:"source_text"`
	TranslatedText string          `json:"translated_text"`
	Language      string           `json:"language"`
	Difficulty    string           `json:"difficulty"`
	SourceType    string           `json:"source_type"`
	Vocabulary    []VocabularyItem `json:"vocabulary"`
	CreatedAt     string           `json:"created_at"`
}

// LessonsHandler handles lesson CRUD operations.
func LessonsHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listLessons(app, w, r)
		case http.MethodPost:
			createLesson(app, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// LessonByIDHandler handles individual lesson operations.
func LessonByIDHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/lessons/")
		if id == "" {
			http.Error(w, "Lesson ID required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getLesson(app, w, r, id)
		case http.MethodDelete:
			deleteLesson(app, w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func createLesson(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request) {
	var req LessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.SourceText == "" || req.TargetLang == "" {
		http.Error(w, "title, source_text, and target_lang are required", http.StatusBadRequest)
		return
	}

	// Translate the text using shared translate logic
	translatedText, vocabulary := TranslateAndExtract(app, req.SourceText, req.TargetLang)

	// Create lesson record
	lessonRecord := models.NewRecord()
	lessonRecord.Set("collection", "lessons")
	lessonRecord.Set("title", req.Title)
	lessonRecord.Set("source_text", req.SourceText)
	lessonRecord.Set("translated_text", translatedText)
	lessonRecord.Set("language", req.TargetLang)
	lessonRecord.Set("difficulty", req.Difficulty)
	lessonRecord.Set("source_type", req.SourceType)
	lessonRecord.Set("created_at", time.Now().UTC().Format(time.RFC3339))

	// Get current user if authenticated
	if authRecord := r.Context().Value("authRecord"); authRecord != nil {
		if ar, ok := authRecord.(*models.Record); ok {
			lessonRecord.Set("user_id", ar.Id)
		}
	}

	if err := app.Dao().SaveRecord(lessonRecord); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create lesson: %v", err), http.StatusInternalServerError)
		return
	}

	// Create lesson_vocabulary junction records
	for i, item := range vocabulary {
		vocabRecord, err := app.Dao().FindFirstRecordByExpr("vocabulary", fmt.Sprintf("word = '%s' AND language = '%s'", item.Word, req.TargetLang))
		if err != nil || vocabRecord == nil {
			// Create vocabulary if not exists
			vocabRecord = models.NewRecord()
			vocabRecord.Set("collection", "vocabulary")
			vocabRecord.Set("word", item.Word)
			vocabRecord.Set("translation", item.Translation)
			vocabRecord.Set("language", req.TargetLang)
			vocabRecord.Set("part_of_speech", item.PartOfSpeech)
			vocabRecord.Set("difficulty", item.Difficulty)
			vocabRecord.Set("example_sentence", item.Example)
			if err := app.Dao().SaveRecord(vocabRecord); err != nil {
				log.Printf("Warning: failed to save vocabulary: %v", err)
				continue
			}
		}

		// Create junction
		junction := models.NewRecord()
		junction.Set("collection", "lesson_vocabulary")
		junction.Set("lesson_id", lessonRecord.Id)
		junction.Set("vocabulary_id", vocabRecord.Id)
		junction.Set("position_in_lesson", i)
		if err := app.Dao().SaveRecord(junction); err != nil {
			log.Printf("Warning: failed to save lesson_vocabulary: %v", err)
		}
	}

	resp := LessonResponse{
		ID:             lessonRecord.Id,
		Title:          lessonRecord.GetString("title"),
		SourceText:     lessonRecord.GetString("source_text"),
		TranslatedText: lessonRecord.GetString("translated_text"),
		Language:       lessonRecord.GetString("language"),
		Difficulty:     lessonRecord.GetString("difficulty"),
		SourceType:     lessonRecord.GetString("source_type"),
		Vocabulary:     vocabulary,
		CreatedAt:      lessonRecord.GetString("created_at"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func listLessons(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	records, err := app.Dao().FindRecordsByExpr("lessons", "", "-created_at", limit, (page-1)*limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list lessons: %v", err), http.StatusInternalServerError)
		return
	}

	var lessons []LessonResponse
	for _, rec := range records {
		// Get vocabulary for this lesson
		vocab := getLessonVocabulary(app, rec.Id)
		lessons = append(lessons, LessonResponse{
			ID:             rec.Id,
			Title:          rec.GetString("title"),
			SourceText:     rec.GetString("source_text"),
			TranslatedText: rec.GetString("translated_text"),
			Language:       rec.GetString("language"),
			Difficulty:     rec.GetString("difficulty"),
			SourceType:     rec.GetString("source_type"),
			Vocabulary:     vocab,
			CreatedAt:      rec.GetString("created_at"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lessons)
}

func getLesson(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, id string) {
	record, err := app.Dao().FindRecordById("lessons", id)
	if err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	vocab := getLessonVocabulary(app, id)

	resp := LessonResponse{
		ID:             record.Id,
		Title:          record.GetString("title"),
		SourceText:     record.GetString("source_text"),
		TranslatedText: record.GetString("translated_text"),
		Language:       record.GetString("language"),
		Difficulty:     record.GetString("difficulty"),
		SourceType:     record.GetString("source_type"),
		Vocabulary:     vocab,
		CreatedAt:      record.GetString("created_at"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func deleteLesson(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, id string) {
	record, err := app.Dao().FindRecordById("lessons", id)
	if err != nil {
		http.Error(w, "Lesson not found", http.StatusNotFound)
		return
	}

	// Delete junction records first
	junctions, _ := app.Dao().FindRecordsByExpr("lesson_vocabulary", fmt.Sprintf("lesson_id = '%s'", id))
	for _, j := range junctions {
		app.Dao().DeleteRecord(j)
	}

	if err := app.Dao().DeleteRecord(record); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete lesson: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getLessonVocabulary(app *pocketbase.PocketBase, lessonID string) []VocabularyItem {
	junctions, err := app.Dao().FindRecordsByExpr("lesson_vocabulary", fmt.Sprintf("lesson_id = '%s'", lessonID), "position_in_lesson")
	if err != nil {
		return nil
	}

	var items []VocabularyItem
	for _, j := range junctions {
		vocabID := j.GetString("vocabulary_id")
		vocab, err := app.Dao().FindRecordById("vocabulary", vocabID)
		if err != nil {
			continue
		}
		items = append(items, VocabularyItem{
			ID:           vocab.Id,
			Word:         vocab.GetString("word"),
			Translation:  vocab.GetString("translation"),
			PartOfSpeech: vocab.GetString("part_of_speech"),
			Difficulty:   vocab.GetString("difficulty"),
			Example:      vocab.GetString("example_sentence"),
		})
	}
	return items
}