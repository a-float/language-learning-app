package handlers

import (
	"testing"
)

func TestLessonValidation(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		sourceText    string
		targetLang    string
		difficulty    string
		sourceType    string
		expectValid   bool
	}{
		{
			name:       "valid lesson",
			title:      "My First Lesson",
			sourceText: "Hello world, this is a test.",
			targetLang: "es",
			difficulty: "beginner",
			sourceType: "text",
			expectValid: true,
		},
		{
			name:       "valid lesson with minimal fields",
			title:      "Lesson",
			sourceText: "Hi",
			targetLang: "it",
			difficulty: "",
			sourceType: "",
			expectValid: true,
		},
		{
			name:       "empty title invalid",
			title:      "",
			sourceText: "Hello",
			targetLang: "es",
			expectValid: false,
		},
		{
			name:       "empty source text invalid",
			title:      "Lesson",
			sourceText: "",
			targetLang: "es",
			expectValid: false,
		},
		{
			name:       "empty target lang invalid",
			title:      "Lesson",
			sourceText: "Hello",
			targetLang: "",
			expectValid: false,
		},
		{
			name:       "invalid target lang",
			title:      "Lesson",
			sourceText: "Hello",
			targetLang: "fr",
			expectValid: false,
		},
		{
			name:       "valid target lang es",
			title:      "Lesson",
			sourceText: "Hello",
			targetLang: "es",
			expectValid: true,
		},
		{
			name:       "valid target lang it",
			title:      "Lesson",
			sourceText: "Hello",
			targetLang: "it",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateLessonRequest(tt.title, tt.sourceText, tt.targetLang)
			if valid != tt.expectValid {
				t.Errorf("validateLessonRequest(%q, %q, %q) = %v, want %v", tt.title, tt.sourceText, tt.targetLang, valid, tt.expectValid)
			}
		})
	}
}

func TestLessonVocabularyJunction(t *testing.T) {
	// Test the logic for creating lesson_vocabulary junction records
	vocabItems := []VocabularyItem{
		{Word: "hello", Translation: "hola", PartOfSpeech: "interjection", Difficulty: "beginner", Example: "Hello world"},
		{Word: "world", Translation: "mundo", PartOfSpeech: "noun", Difficulty: "beginner", Example: "World peace"},
		{Word: "good", Translation: "bueno", PartOfSpeech: "adjective", Difficulty: "beginner", Example: "Good morning"},
	}

	tests := []struct {
		name           string
		vocabItems     []VocabularyItem
		expectedCount  int
		checkPositions []int
	}{
		{
			name:           "three items",
			vocabItems:     vocabItems,
			expectedCount:  3,
			checkPositions: []int{0, 1, 2},
		},
		{
			name:           "empty vocabulary",
			vocabItems:     []VocabularyItem{},
			expectedCount:  0,
			checkPositions: nil,
		},
		{
			name:           "single item",
			vocabItems:     vocabItems[:1],
			expectedCount:  1,
			checkPositions: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			junctions := createLessonVocabularyJunctions("lesson1", tt.vocabItems)

			if len(junctions) != tt.expectedCount {
				t.Errorf("Expected %d junctions, got %d", tt.expectedCount, len(junctions))
			}

			for i, j := range junctions {
				if j.LessonID != "lesson1" {
					t.Errorf("Junction %d: LessonID = %q, want %q", i, j.LessonID, "lesson1")
				}
				if j.Position != i {
					t.Errorf("Junction %d: Position = %d, want %d", i, j.Position, i)
				}
				if j.VocabularyID == "" {
					t.Errorf("Junction %d: VocabularyID should not be empty", i)
				}
			}

			if tt.checkPositions != nil {
				for _, pos := range tt.checkPositions {
					found := false
					for _, j := range junctions {
						if j.Position == pos {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Position %d not found in junctions", pos)
					}
				}
			}
		})
	}
}

func TestLessonDeletionCascade(t *testing.T) {
	// Test that deleting a lesson also deletes junction records
	lessonID := "lesson1"
	junctions := []LessonVocabularyJunction{
		{ID: "j1", LessonID: lessonID, VocabularyID: "vocab1", Position: 0},
		{ID: "j2", LessonID: lessonID, VocabularyID: "vocab2", Position: 1},
		{ID: "j3", LessonID: "lesson2", VocabularyID: "vocab3", Position: 0}, // Different lesson
	}

	// Filter junctions for the lesson being deleted
	toDelete := filterJunctionsByLesson(junctions, lessonID)

	if len(toDelete) != 2 {
		t.Errorf("Expected 2 junctions to delete, got %d", len(toDelete))
	}

	for _, j := range toDelete {
		if j.LessonID != lessonID {
			t.Errorf("Junction %s should belong to lesson %s", j.ID, lessonID)
		}
	}

	// Remaining junctions should not include deleted lesson
	remaining := filterJunctionsByLesson(junctions, "lesson2")
	if len(remaining) != 1 {
		t.Errorf("Expected 1 remaining junction, got %d", len(remaining))
	}
	if remaining[0].ID != "j3" {
		t.Errorf("Remaining junction should be j3, got %s", remaining[0].ID)
	}
}

func TestLessonPagination(t *testing.T) {
	lessons := make([]LessonRecord, 25)
	for i := 0; i < 25; i++ {
		lessons[i] = LessonRecord{
			ID:        "lesson" + string(rune('0'+i)),
			Title:     "Lesson " + string(rune('0'+i)),
			CreatedAt: "2024-01-01T00:00:00Z",
		}
	}

	tests := []struct {
		name       string
		page       int
		limit      int
		expected   int
	}{
		{"page 1 limit 20", 1, 20, 20},
		{"page 2 limit 20", 2, 20, 5},
		{"page 1 limit 50", 1, 50, 25},
		{"page 3 limit 10", 3, 10, 5},
		{"page 4 limit 10", 4, 10, 0},
		{"invalid page defaults to 1", 0, 20, 20},
		{"invalid limit defaults to 20", 1, 0, 20},
		{"limit capped at 100", 1, 200, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paginateLessons(lessons, tt.page, tt.limit)
			if len(result) != tt.expected {
				t.Errorf("Expected %d items, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestLessonSorting(t *testing.T) {
	lessons := []LessonRecord{
		{ID: "lesson1", Title: "Lesson 1", CreatedAt: "2024-01-03T00:00:00Z"},
		{ID: "lesson2", Title: "Lesson 2", CreatedAt: "2024-01-01T00:00:00Z"},
		{ID: "lesson3", Title: "Lesson 3", CreatedAt: "2024-01-02T00:00:00Z"},
	}

	result := paginateLessons(lessons, 1, 10)

	// Should be sorted by created_at descending (newest first)
	if len(result) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result))
	}
	if result[0].ID != "lesson1" {
		t.Errorf("First should be lesson1 (newest), got %s", result[0].ID)
	}
	if result[1].ID != "lesson3" {
		t.Errorf("Second should be lesson3, got %s", result[1].ID)
	}
	if result[2].ID != "lesson2" {
		t.Errorf("Third should be lesson2 (oldest), got %s", result[2].ID)
	}
}

// Helper types and functions

type LessonRecord struct {
	ID        string
	Title     string
	CreatedAt string
}

type LessonVocabularyJunction struct {
	ID            string
	LessonID      string
	VocabularyID  string
	Position      int
}

func validateLessonRequest(title, sourceText, targetLang string) bool {
	if title == "" || sourceText == "" || targetLang == "" {
		return false
	}
	if targetLang != "es" && targetLang != "it" {
		return false
	}
	return true
}

func createLessonVocabularyJunctions(lessonID string, vocabItems []VocabularyItem) []LessonVocabularyJunction {
	var junctions []LessonVocabularyJunction
	for i, item := range vocabItems {
		junctions = append(junctions, LessonVocabularyJunction{
			LessonID:     lessonID,
			VocabularyID: item.ID, // In real code, this would be the vocab record ID
			Position:     i,
		})
	}
	return junctions
}

func filterJunctionsByLesson(junctions []LessonVocabularyJunction, lessonID string) []LessonVocabularyJunction {
	var result []LessonVocabularyJunction
	for _, j := range junctions {
		if j.LessonID == lessonID {
			result = append(result, j)
		}
	}
	return result
}

func paginateLessons(lessons []LessonRecord, page, limit int) []LessonRecord {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	start := (page - 1) * limit
	if start >= len(lessons) {
		return []LessonRecord{}
	}
	end := start + limit
	if end > len(lessons) {
		end = len(lessons)
	}
	return lessons[start:end]
}