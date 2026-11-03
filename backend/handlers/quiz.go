package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// QuizQuestion represents a single quiz question.
type QuizQuestion struct {
	ID            string   `json:"id"`
	VocabularyID  string   `json:"vocabulary_id"`
	Question      string   `json:"question"`       // The word to translate
	QuestionLang  string   `json:"question_lang"`  // Source language (always "en")
	Answer        string   `json:"answer"`         // Correct translation
	AnswerLang    string   `json:"answer_lang"`    // Target language (es/it)
	Options       []string `json:"options"`        // Multiple choice options
	QuestionType  string   `json:"question_type"`  // "multiple_choice", "type_answer"
	Direction     string   `json:"direction"`      // "en_to_target" or "target_to_en"
}

// QuizRequest represents a quiz generation request.
type QuizRequest struct {
	Source       string `json:"source"`        // "favorites", "progress", "lesson", "mixed"
	SourceID     string `json:"source_id,omitempty"` // Lesson ID if source="lesson"
	QuestionCount int   `json:"question_count"` // Number of questions (default 10)
	QuestionType  string `json:"question_type,omitempty"` // "multiple_choice", "type_answer", "mixed"
	Direction     string `json:"direction,omitempty"` // "en_to_target", "target_to_en", "mixed"
	TargetLang    string `json:"target_lang,omitempty"` // Filter by language (es/it)
}

// QuizResponse represents a generated quiz.
type QuizResponse struct {
	Questions []QuizQuestion `json:"questions"`
	Total     int            `json:"total"`
}

// QuizSubmitRequest represents a quiz submission.
type QuizSubmitRequest struct {
	Answers []QuizAnswer `json:"answers"`
}

// QuizAnswer represents a single answer submission.
type QuizAnswer struct {
	QuestionID   string `json:"question_id"`
	UserAnswer   string `json:"user_answer"`
	IsCorrect    bool   `json:"is_correct"`
	TimeTakenMs  int    `json:"time_taken_ms,omitempty"`
}

// QuizResultResponse represents quiz results.
type QuizResultResponse struct {
	Score           int            `json:"score"`            // Correct answers count
	Total           int            `json:"total"`            // Total questions
	Percentage      float64        `json:"percentage"`       // Score percentage
	Results         []QuizResult   `json:"results"`          // Per-question results
	XPEarned        int            `json:"xp_earned"`        // Experience points earned
}

// QuizResult represents a single question result.
type QuizResult struct {
	QuestionID   string `json:"question_id"`
	VocabularyID string `json:"vocabulary_id"`
	UserAnswer   string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	IsCorrect    bool   `json:"is_correct"`
	TimeTakenMs  int    `json:"time_taken_ms,omitempty"`
}

// QuizHandler handles quiz generation and submission.
func QuizHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authRecord := r.Context().Value("authRecord")
		if authRecord == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID := authRecord.(*models.Record).Id

		switch r.Method {
		case http.MethodPost:
			// Check if it's quiz generation or submission
			path := r.URL.Path
			if strings.HasSuffix(path, "/submit") {
				submitQuiz(app, w, r, userID)
			} else {
				generateQuiz(app, w, r, userID)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func generateQuiz(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID string) {
	var req QuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Defaults
	if req.QuestionCount <= 0 {
		req.QuestionCount = 10
	}
	if req.QuestionCount > 50 {
		req.QuestionCount = 50
	}
	if req.Source == "" {
		req.Source = "mixed"
	}

	// Get vocabulary pool based on source
	vocabPool, err := getVocabularyPool(app, userID, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get vocabulary pool: %v", err), http.StatusInternalServerError)
		return
	}

	if len(vocabPool) == 0 {
		http.Error(w, "No vocabulary available for quiz. Add some words first!", http.StatusBadRequest)
		return
	}

	// Generate questions
	questions := generateQuestions(vocabPool, req.QuestionCount, req.QuestionType, req.Direction, req.TargetLang)

	resp := QuizResponse{
		Questions: questions,
		Total:     len(questions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func submitQuiz(app *pocketbase.PocketBase, w http.ResponseWriter, r *http.Request, userID string) {
	var req QuizSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.Answers) == 0 {
		http.Error(w, "No answers provided", http.StatusBadRequest)
		return
	}

	var results []QuizResult
	score := 0
	totalXP := 0

	for _, ans := range req.Answers {
		// Get the question details to verify
		// In a real implementation, you'd store the quiz session or reconstruct from question ID
		// For now, we'll trust the client's is_correct but verify vocabulary exists
		vocab, err := app.Dao().FindRecordById("vocabulary", ans.QuestionID)
		if err != nil {
			continue
		}

		correctAnswer := vocab.GetString("translation")
		isCorrect := strings.EqualFold(strings.TrimSpace(ans.UserAnswer), strings.TrimSpace(correctAnswer))

		// If it's multiple choice, the questionID might be the vocab ID
		// For simplicity, we'll check both word and translation
		if !isCorrect {
			// Check if answer matches the word (reverse direction)
			isCorrect = strings.EqualFold(strings.TrimSpace(ans.UserAnswer), strings.TrimSpace(vocab.GetString("word")))
		}

		if isCorrect {
			score++
			totalXP += 10 // Base XP per correct answer
			// Bonus XP for speed
			if ans.TimeTakenMs > 0 && ans.TimeTakenMs < 5000 {
				totalXP += 5
			}
		}

		// Update progress for this vocabulary
		updateQuizProgress(app, userID, vocab.Id, isCorrect)

		results = append(results, QuizResult{
			QuestionID:    ans.QuestionID,
			VocabularyID:  vocab.Id,
			UserAnswer:    ans.UserAnswer,
			CorrectAnswer: correctAnswer,
			IsCorrect:     isCorrect,
			TimeTakenMs:   ans.TimeTakenMs,
		})
	}

	percentage := 0.0
	if len(req.Answers) > 0 {
		percentage = float64(score) / float64(len(req.Answers)) * 100
	}

	resp := QuizResultResponse{
		Score:      score,
		Total:      len(req.Answers),
		Percentage: percentage,
		Results:    results,
		XPEarned:   totalXP,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getVocabularyPool(app *pocketbase.PocketBase, userID string, req QuizRequest) ([]*models.Record, error) {
	var filter string
	var records []*models.Record
	var err error

	switch req.Source {
	case "favorites":
		// Get vocabulary IDs from user's favorites
		favs, err := app.Dao().FindRecordsByExpr("favorites", fmt.Sprintf("user_id = '%s'", userID))
		if err != nil {
			return nil, err
		}
		var vocabIDs []string
		for _, f := range favs {
			vocabIDs = append(vocabIDs, f.GetString("vocabulary_id"))
		}
		if len(vocabIDs) == 0 {
			return []*models.Record{}, nil
		}
		filter = fmt.Sprintf("id IN ('%s')", strings.Join(vocabIDs, "','"))

	case "progress":
		// Get vocabulary from user's progress (words they've studied)
		prog, err := app.Dao().FindRecordsByExpr("user_vocabulary_progress", fmt.Sprintf("user_id = '%s'", userID))
		if err != nil {
			return nil, err
		}
		var vocabIDs []string
		for _, p := range prog {
			vocabIDs = append(vocabIDs, p.GetString("vocabulary_id"))
		}
		if len(vocabIDs) == 0 {
			return []*models.Record{}, nil
		}
		filter = fmt.Sprintf("id IN ('%s')", strings.Join(vocabIDs, "','"))

	case "lesson":
		if req.SourceID == "" {
			return nil, fmt.Errorf("source_id (lesson ID) required for lesson source")
		}
		// Get vocabulary from specific lesson
		junctions, err := app.Dao().FindRecordsByExpr("lesson_vocabulary", fmt.Sprintf("lesson_id = '%s'", req.SourceID))
		if err != nil {
			return nil, err
		}
		var vocabIDs []string
		for _, j := range junctions {
			vocabIDs = append(vocabIDs, j.GetString("vocabulary_id"))
		}
		if len(vocabIDs) == 0 {
			return []*models.Record{}, nil
		}
		filter = fmt.Sprintf("id IN ('%s')", strings.Join(vocabIDs, "','"))

	case "mixed":
		fallthrough
	default:
		// All vocabulary, optionally filtered by language
		filter = ""
		if req.TargetLang != "" {
			filter = fmt.Sprintf("language = '%s'", req.TargetLang)
		}
	}

	records, err = app.Dao().FindRecordsByExpr("vocabulary", filter)
	return records, err
}

func generateQuestions(vocabPool []*models.Record, count int, questionType, direction, targetLang string) []QuizQuestion {
	rand.Seed(time.Now().UnixNano())
	
	// Shuffle pool
	rand.Shuffle(len(vocabPool), func(i, j int) {
		vocabPool[i], vocabPool[j] = vocabPool[j], vocabPool[i]
	})

	// Take up to count items
	if len(vocabPool) > count {
		vocabPool = vocabPool[:count]
	}

	var questions []QuizQuestion
	for _, vocab := range vocabPool {
		word := vocab.GetString("word")
		translation := vocab.GetString("translation")
		lang := vocab.GetString("language")
		vocabID := vocab.Id

		// Determine direction for this question
		qDirection := direction
		if qDirection == "mixed" || qDirection == "" {
			if rand.Float32() < 0.5 {
				qDirection = "en_to_target"
			} else {
				qDirection = "target_to_en"
			}
		}

		// Determine question type
		qType := questionType
		if qType == "mixed" || qType == "" {
			if rand.Float32() < 0.7 {
				qType = "multiple_choice"
			} else {
				qType = "type_answer"
			}
		}

		var question, answer, qLang, aLang string
		var options []string

		if qDirection == "en_to_target" {
			question = word
			answer = translation
			qLang = "en"
			aLang = lang
		} else {
			question = translation
			answer = word
			qLang = lang
			aLang = "en"
		}

		if qType == "multiple_choice" {
			// Generate 3 distractors + correct answer
			options = generateOptions(vocabPool, vocab, answer, qDirection, 4)
		}

		questions = append(questions, QuizQuestion{
			ID:           vocabID,
			VocabularyID: vocabID,
			Question:     question,
			QuestionLang: qLang,
			Answer:       answer,
			AnswerLang:   aLang,
			Options:      options,
			QuestionType: qType,
			Direction:    qDirection,
		})
	}

	return questions
}

func generateOptions(vocabPool []*models.Record, correctVocab *models.Record, correctAnswer, direction string, numOptions int) []string {
	options := []string{correctAnswer}
	
	// Collect distractors from other vocabulary items
	var distractors []string
	for _, v := range vocabPool {
		if v.Id == correctVocab.Id {
			continue
		}
		if direction == "en_to_target" {
			distractors = append(distractors, v.GetString("translation"))
		} else {
			distractors = append(distractors, v.GetString("word"))
		}
		if len(distractors) >= numOptions-1 {
			break
		}
	}

	// If not enough distractors, add some generic ones
	for len(distractors) < numOptions-1 {
		distractors = append(distractors, fmt.Sprintf("option_%d", len(distractors)+1))
	}

	// Shuffle and take needed amount
	rand.Shuffle(len(distractors), func(i, j int) {
		distractors[i], distractors[j] = distractors[j], distractors[i]
	})
	options = append(options, distractors[:numOptions-1]...)

	// Shuffle final options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	return options
}

func updateQuizProgress(app *pocketbase.PocketBase, userID, vocabID string, correct bool) {
	record, err := app.Dao().FindFirstRecordByExpr("user_vocabulary_progress", fmt.Sprintf("user_id = '%s' && vocabulary_id = '%s'", userID, vocabID))
	now := time.Now().UTC().Format(time.RFC3339)

	if err != nil || record == nil {
		record = models.NewRecord()
		record.Set("collection", "user_vocabulary_progress")
		record.Set("user_id", userID)
		record.Set("vocabulary_id", vocabID)
		record.Set("knowledge_level", 0)
		record.Set("last_reviewed", now)
		record.Set("review_count", 1)
	} else {
		record.Set("last_reviewed", now)
		record.Set("review_count", record.GetInt("review_count")+1)
		
		// Adjust knowledge level based on correctness
		currentLevel := int(record.GetInt("knowledge_level"))
		if correct {
			if currentLevel < 5 {
				record.Set("knowledge_level", currentLevel+1)
			}
		} else {
			if currentLevel > 0 {
				record.Set("knowledge_level", currentLevel-1)
			}
		}
	}

	if err := app.Dao().SaveRecord(record); err != nil {
		log.Printf("Warning: failed to update quiz progress: %v", err)
	}
}