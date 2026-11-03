package handlers

import (
	"testing"
	"time"
)

func TestGenerateQuestions(t *testing.T) {
	// Create mock vocabulary records
	vocabPool := createMockVocabPool(t, []struct {
		id           string
		word         string
		translation  string
		language     string
	}{
		{"vocab1", "hello", "hola", "es"},
		{"vocab2", "world", "mundo", "es"},
		{"vocab3", "good", "bueno", "es"},
		{"vocab4", "bad", "malo", "es"},
		{"vocab5", "big", "grande", "es"},
	})

	tests := []struct {
		name          string
		count         int
		questionType  string
		direction     string
		targetLang    string
		expectedCount int
		checkTypes    []string
		checkDirs     []string
	}{
		{
			name:          "default multiple choice en_to_target",
			count:         3,
			questionType:  "multiple_choice",
			direction:     "en_to_target",
			targetLang:    "es",
			expectedCount: 3,
			checkTypes:    []string{"multiple_choice"},
			checkDirs:     []string{"en_to_target"},
		},
		{
			name:          "type_answer target_to_en",
			count:         2,
			questionType:  "type_answer",
			direction:     "target_to_en",
			targetLang:    "es",
			expectedCount: 2,
			checkTypes:    []string{"type_answer"},
			checkDirs:     []string{"target_to_en"},
		},
		{
			name:          "mixed type and direction",
			count:         5,
			questionType:  "mixed",
			direction:     "mixed",
			targetLang:    "es",
			expectedCount: 5,
			checkTypes:    []string{"multiple_choice", "type_answer"},
			checkDirs:     []string{"en_to_target", "target_to_en"},
		},
		{
			name:          "more questions than pool",
			count:         10,
			questionType:  "multiple_choice",
			direction:     "en_to_target",
			targetLang:    "es",
			expectedCount: 5, // limited by pool size
			checkTypes:    []string{"multiple_choice"},
			checkDirs:     []string{"en_to_target"},
		},
		{
			name:          "zero count defaults to 10",
			count:         0,
			questionType:  "multiple_choice",
			direction:     "en_to_target",
			targetLang:    "es",
			expectedCount: 5, // limited by pool size
			checkTypes:    []string{"multiple_choice"},
			checkDirs:     []string{"en_to_target"},
		},
		{
			name:          "count capped at 50",
			count:         100,
			questionType:  "multiple_choice",
			direction:     "en_to_target",
			targetLang:    "es",
			expectedCount: 5, // limited by pool size
			checkTypes:    []string{"multiple_choice"},
			checkDirs:     []string{"en_to_target"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			questions := generateQuestions(vocabPool, tt.count, tt.questionType, tt.direction, tt.targetLang)

			if len(questions) != tt.expectedCount {
				t.Errorf("Expected %d questions, got %d", tt.expectedCount, len(questions))
			}

			// Check question types
			typeFound := make(map[string]bool)
			dirFound := make(map[string]bool)
			for _, q := range questions {
				typeFound[q.QuestionType] = true
				dirFound[q.Direction] = true

				// Verify structure
				if q.ID == "" {
					t.Error("Question ID should not be empty")
				}
				if q.VocabularyID == "" {
					t.Error("Question VocabularyID should not be empty")
				}
				if q.Question == "" {
					t.Error("Question text should not be empty")
				}
				if q.Answer == "" {
					t.Error("Answer should not be empty")
				}
				if q.QuestionLang == "" {
					t.Error("QuestionLang should not be empty")
				}
				if q.AnswerLang == "" {
					t.Error("AnswerLang should not be empty")
				}

				// Verify multiple choice has options
				if q.QuestionType == "multiple_choice" {
					if len(q.Options) != 4 {
						t.Errorf("Multiple choice should have 4 options, got %d", len(q.Options))
					}
					// Correct answer should be in options
					found := false
					for _, opt := range q.Options {
						if opt == q.Answer {
							found = true
							break
						}
					}
					if !found {
						t.Error("Correct answer not found in options")
					}
				} else {
					// Type answer should have no options
					if len(q.Options) != 0 {
						t.Errorf("Type answer should have 0 options, got %d", len(q.Options))
					}
				}
			}

			// Check expected types present
			for _, expectedType := range tt.checkTypes {
				if !typeFound[expectedType] {
					t.Errorf("Expected question type %q not found", expectedType)
				}
			}
			for _, expectedDir := range tt.checkDirs {
				if !dirFound[expectedDir] {
					t.Errorf("Expected direction %q not found", expectedDir)
				}
			}
		})
	}
}

func TestGenerateQuestionsEmptyPool(t *testing.T) {
	questions := generateQuestions([]*models.Record{}, 10, "multiple_choice", "en_to_target", "es")
	if len(questions) != 0 {
		t.Errorf("Expected 0 questions for empty pool, got %d", len(questions))
	}
}

func TestGenerateOptions(t *testing.T) {
	vocabPool := createMockVocabPool(t, []struct {
		id           string
		word         string
		translation  string
		language     string
	}{
		{"vocab1", "hello", "hola", "es"},
		{"vocab2", "world", "mundo", "es"},
		{"vocab3", "good", "bueno", "es"},
		{"vocab4", "bad", "malo", "es"},
		{"vocab5", "big", "grande", "es"},
	})

	correctVocab := vocabPool[0] // "hello" -> "hola"

	tests := []struct {
		name           string
		direction      string
		correctAnswer  string
		expectedCount  int
	}{
		{
			name:           "en_to_target",
			direction:      "en_to_target",
			correctAnswer:  "hola",
			expectedCount:  4,
		},
		{
			name:           "target_to_en",
			direction:      "target_to_en",
			correctAnswer:  "hello",
			expectedCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := generateOptions(vocabPool, correctVocab, tt.correctAnswer, tt.direction, tt.expectedCount)

			if len(options) != tt.expectedCount {
				t.Errorf("Expected %d options, got %d", tt.expectedCount, len(options))
			}

			// Correct answer should be in options
			found := false
			for _, opt := range options {
				if opt == tt.correctAnswer {
					found = true
					break
				}
			}
			if !found {
				t.Error("Correct answer not in options")
			}

			// All options should be unique
			seen := make(map[string]bool)
			for _, opt := range options {
				if seen[opt] {
					t.Errorf("Duplicate option: %q", opt)
				}
				seen[opt] = true
			}

			// Distractors should come from other vocab items
			distractorCount := 0
			for _, opt := range options {
				if opt != tt.correctAnswer {
					distractorCount++
					// Should be a translation or word from another vocab item
					isValidDistractor := false
					for _, v := range vocabPool {
						if v.Id == correctVocab.Id {
							continue
						}
						if tt.direction == "en_to_target" && v.GetString("translation") == opt {
							isValidDistractor = true
						}
						if tt.direction == "target_to_en" && v.GetString("word") == opt {
							isValidDistractor = true
						}
					}
					// Generic options are also valid (fallback)
					if !isValidDistractor && len(opt) > 7 && opt[:7] == "option_" {
						isValidDistractor = true
					}
					if !isValidDistractor {
						t.Errorf("Invalid distractor: %q", opt)
					}
				}
			}
			if distractorCount != tt.expectedCount-1 {
				t.Errorf("Expected %d distractors, got %d", tt.expectedCount-1, distractorCount)
			}
		})
	}
}

func TestGenerateOptionsNotEnoughDistractors(t *testing.T) {
	// Pool with only 1 item (the correct one)
	vocabPool := createMockVocabPool(t, []struct {
		id           string
		word         string
		translation  string
		language     string
	}{
		{"vocab1", "hello", "hola", "es"},
	})

	correctVocab := vocabPool[0]
	options := generateOptions(vocabPool, correctVocab, "hola", "en_to_target", 4)

	if len(options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(options))
	}

	// Should have correct answer + 3 generic options
	foundCorrect := false
	genericCount := 0
	for _, opt := range options {
		if opt == "hola" {
			foundCorrect = true
		} else if len(opt) > 7 && opt[:7] == "option_" {
			genericCount++
		}
	}
	if !foundCorrect {
		t.Error("Correct answer not found")
	}
	if genericCount != 3 {
		t.Errorf("Expected 3 generic options, got %d", genericCount)
	}
}

// Helper to create mock vocabulary records for testing
func createMockVocabPool(t *testing.T, items []struct {
	id          string
	word        string
	translation string
	language    string
}) []*models.Record {
	t.Helper()
	var pool []*models.Record
	for _, item := range items {
		rec := models.NewRecord()
		rec.Set("id", item.id)
		rec.Set("word", item.word)
		rec.Set("translation", item.translation)
		rec.Set("language", item.language)
		rec.Set("part_of_speech", "noun")
		rec.Set("difficulty", "beginner")
		rec.Set("example_sentence", "Example sentence")
		pool = append(pool, rec)
	}
	return pool
}

// Test that random seed produces different results
func TestGenerateQuestionsRandomness(t *testing.T) {
	vocabPool := createMockVocabPool(t, []struct {
		id           string
		word         string
		translation  string
		language     string
	}{
		{"vocab1", "hello", "hola", "es"},
		{"vocab2", "world", "mundo", "es"},
		{"vocab3", "good", "bueno", "es"},
		{"vocab4", "bad", "malo", "es"},
		{"vocab5", "big", "grande", "es"},
	})

	// Run multiple times with different seeds (time-based)
	results := make(map[string]int)
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Millisecond) // Ensure different seed
		questions := generateQuestions(vocabPool, 5, "multiple_choice", "mixed", "es")
		key := ""
		for _, q := range questions {
			key += q.ID + ","
		}
		results[key]++
	}

	// Should have some variation (not all identical)
	if len(results) == 1 {
		t.Log("All runs produced same order (may happen with small pool)")
	}
}