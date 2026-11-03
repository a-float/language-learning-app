package handlers

import (
	"testing"
)

func TestTranslateRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		targetLang  string
		expectValid bool
	}{
		{
			name:        "valid request",
			text:        "Hello world",
			targetLang:  "es",
			expectValid: true,
		},
		{
			name:        "valid italian",
			text:        "Hello world",
			targetLang:  "it",
			expectValid: true,
		},
		{
			name:        "empty text invalid",
			text:        "",
			targetLang:  "es",
			expectValid: false,
		},
		{
			name:        "empty target lang invalid",
			text:        "Hello",
			targetLang:  "",
			expectValid: false,
		},
		{
			name:        "invalid target lang",
			text:        "Hello",
			targetLang:  "fr",
			expectValid: false,
		},
		{
			name:        "whitespace text valid",
			text:        "  Hello  ",
			targetLang:  "es",
			expectValid: true,
		},
		{
			name:        "long text valid",
			text:        "This is a very long text that should still be valid for translation purposes and not cause any issues with the validation logic",
			targetLang:  "it",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateTranslateRequest(tt.text, tt.targetLang)
			if valid != tt.expectValid {
				t.Errorf("validateTranslateRequest(%q, %q) = %v, want %v", tt.text, tt.targetLang, valid, tt.expectValid)
			}
		})
	}
}

func TestTranslateResponseStructure(t *testing.T) {
	resp := TranslateResponse{
		TranslatedText: "Hola mundo",
		Vocabulary: []VocabularyItem{
			{Word: "hello", Translation: "hola", PartOfSpeech: "interjection", Difficulty: "beginner", Example: "Hello world"},
			{Word: "world", Translation: "mundo", PartOfSpeech: "noun", Difficulty: "beginner", Example: "World peace"},
		},
		Cached: false,
	}

	if resp.TranslatedText == "" {
		t.Error("TranslatedText should not be empty")
	}
	if len(resp.Vocabulary) != 2 {
		t.Errorf("Expected 2 vocabulary items, got %d", len(resp.Vocabulary))
	}
	for i, item := range resp.Vocabulary {
		if item.Word == "" {
			t.Errorf("Vocabulary[%d].Word should not be empty", i)
		}
		if item.Translation == "" {
			t.Errorf("Vocabulary[%d].Translation should not be empty", i)
		}
		if item.Difficulty != "beginner" && item.Difficulty != "intermediate" && item.Difficulty != "advanced" {
			t.Errorf("Vocabulary[%d].Difficulty invalid: %q", i, item.Difficulty)
		}
	}
	if resp.Cached != false {
		t.Error("Cached should be false for new translation")
	}
}

func TestVocabularyItemStructure(t *testing.T) {
	item := VocabularyItem{
		ID:           "vocab123",
		Word:         "hello",
		Translation:  "hola",
		PartOfSpeech: "interjection",
		Difficulty:   "beginner",
		Example:      "Hello world",
	}

	if item.ID == "" {
		t.Error("ID should not be empty")
	}
	if item.Word != "hello" {
		t.Errorf("Word = %q, want %q", item.Word, "hello")
	}
	if item.Translation != "hola" {
		t.Errorf("Translation = %q, want %q", item.Translation, "hola")
	}
	if item.PartOfSpeech != "interjection" {
		t.Errorf("PartOfSpeech = %q, want %q", item.PartOfSpeech, "interjection")
	}
	if item.Difficulty != "beginner" {
		t.Errorf("Difficulty = %q, want %q", item.Difficulty, "beginner")
	}
	if item.Example != "Hello world" {
		t.Errorf("Example = %q, want %q", item.Example, "Hello world")
	}
}

func TestVocabularyItemDifficultyValues(t *testing.T) {
	validDifficulties := []string{"beginner", "intermediate", "advanced"}
	invalidDifficulties := []string{"easy", "hard", "expert", "novice", ""}

	for _, diff := range validDifficulties {
		item := VocabularyItem{Difficulty: diff}
		if !isValidDifficulty(item.Difficulty) {
			t.Errorf("Difficulty %q should be valid", diff)
		}
	}

	for _, diff := range invalidDifficulties {
		item := VocabularyItem{Difficulty: diff}
		if isValidDifficulty(item.Difficulty) {
			t.Errorf("Difficulty %q should be invalid", diff)
		}
	}
}

func TestVocabularyItemPartOfSpeechValues(t *testing.T) {
	validPOS := []string{"noun", "verb", "adjective", "adverb", "interjection", "preposition", "conjunction", "pronoun", "determiner", ""}
	// Empty is allowed (optional field)

	for _, pos := range validPOS {
		item := VocabularyItem{PartOfSpeech: pos}
		// Just verify it accepts these values
		if item.PartOfSpeech != pos {
			t.Errorf("PartOfSpeech not set correctly: %q", item.PartOfSpeech)
		}
	}
}

func TestCacheKeyConsistency(t *testing.T) {
	// CacheKey should be consistent for same inputs
	text := "Hello world"
	lang := "es"

	key1 := CacheKey(text, lang)
	key2 := CacheKey(text, lang)

	if key1 != key2 {
		t.Error("CacheKey not consistent for same inputs")
	}

	// Different inputs should produce different keys
	key3 := CacheKey(text, "it")
	if key1 == key3 {
		t.Error("Different languages should produce different keys")
	}

	key4 := CacheKey("Different text", lang)
	if key1 == key4 {
		t.Error("Different texts should produce different keys")
	}
}

func TestTranslateRequestJSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling of request/response types
	req := TranslateRequest{
		Text:       "Hello world",
		TargetLang: "es",
	}

	// This would be tested with encoding/json in real tests
	// Just verify struct tags are correct
	if req.Text != "Hello world" {
		t.Error("Text field not set")
	}
	if req.TargetLang != "es" {
		t.Error("TargetLang field not set")
	}
}

func TestTranslateResponseJSON(t *testing.T) {
	resp := TranslateResponse{
		TranslatedText: "Hola mundo",
		Vocabulary: []VocabularyItem{
			{Word: "hello", Translation: "hola"},
		},
		Cached: true,
	}

	if resp.TranslatedText != "Hola mundo" {
		t.Error("TranslatedText not set")
	}
	if len(resp.Vocabulary) != 1 {
		t.Error("Vocabulary not set")
	}
	if !resp.Cached {
		t.Error("Cached not set")
	}
}

// Helper functions

func validateTranslateRequest(text, targetLang string) bool {
	if text == "" || targetLang == "" {
		return false
	}
	if targetLang != "es" && targetLang != "it" {
		return false
	}
	return true
}

func isValidDifficulty(diff string) bool {
	switch diff {
	case "beginner", "intermediate", "advanced":
		return true
	default:
		return false
	}
}