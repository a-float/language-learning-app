package handlers

import (
	"testing"
)

func TestExtractVocabularyFallback(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		translation string
		targetLang string
		minItems   int
		maxItems   int
		checkWords []string
	}{
		{
			name:        "simple sentence",
			source:      "Hello world",
			translation: "Hola mundo",
			targetLang:  "es",
			minItems:    2,
			maxItems:    2,
			checkWords:  []string{"Hello", "world"},
		},
		{
			name:        "sentence with punctuation",
			source:      "Hello, world! How are you?",
			translation: "Hola, mundo! ¿Cómo estás?",
			targetLang:  "es",
			minItems:    4,
			maxItems:    5,
			checkWords:  []string{"Hello", "world", "How", "are", "you"},
		},
		{
			name:        "short words filtered out",
			source:      "I am a boy",
			translation: "Soy un niño",
			targetLang:  "es",
			minItems:    1,
			maxItems:    2,
			checkWords:  []string{"boy"}, // "I", "am", "a" are < 3 chars
		},
		{
			name:        "duplicate words deduplicated",
			source:      "Hello hello world world",
			translation: "Hola hola mundo mundo",
			targetLang:  "es",
			minItems:    2,
			maxItems:    2,
			checkWords:  []string{"Hello", "world"},
		},
		{
			name:        "case insensitive deduplication",
			source:      "Hello HELLO hello",
			translation: "Hola HOLA hola",
			targetLang:  "es",
			minItems:    1,
			maxItems:    1,
			checkWords:  []string{"Hello"},
		},
		{
			name:        "empty source",
			source:      "",
			translation: "",
			targetLang:  "es",
			minItems:    0,
			maxItems:    0,
			checkWords:  nil,
		},
		{
			name:        "only short words",
			source:      "a an the is",
			translation: "un el la es",
			targetLang:  "es",
			minItems:    0,
			maxItems:    0,
			checkWords:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVocabularyFallback(tt.source, tt.translation, tt.targetLang)

			if len(result) < tt.minItems {
				t.Errorf("Expected at least %d items, got %d: %v", tt.minItems, len(result), result)
			}
			if len(result) > tt.maxItems {
				t.Errorf("Expected at most %d items, got %d: %v", tt.maxItems, len(result), result)
			}

			if tt.checkWords != nil {
				found := make(map[string]bool)
				for _, item := range result {
					found[item.Word] = true
				}
				for _, word := range tt.checkWords {
					if !found[word] {
						t.Errorf("Expected word %q not found in result: %v", word, result)
					}
				}
			}

			// Verify all items have required fields
			for _, item := range result {
				if item.Word == "" {
					t.Error("VocabularyItem.Word should not be empty")
				}
				if item.Difficulty == "" {
					t.Error("VocabularyItem.Difficulty should not be empty")
				}
				if item.Difficulty != "beginner" && item.Difficulty != "intermediate" && item.Difficulty != "advanced" {
					t.Errorf("Unexpected difficulty: %q", item.Difficulty)
				}
			}
		})
	}
}

func TestExtractVocabularyFallbackWordCleaning(t *testing.T) {
	source := "Hello, world! (test) [example] {braces}"
	result := extractVocabularyFallback(source, "translation", "es")

	found := make(map[string]bool)
	for _, item := range result {
		found[item.Word] = true
	}

	// Words should be cleaned of punctuation
	expectedWords := []string{"Hello", "world", "test", "example", "braces"}
	for _, word := range expectedWords {
		if !found[word] {
			t.Errorf("Expected cleaned word %q not found in result: %v", word, result)
		}
	}

	// Should not contain punctuation
	for _, item := range result {
		for _, ch := range ",.!?;:\"()[]{}" {
			if containsRune(item.Word, ch) {
				t.Errorf("Word %q should not contain punctuation %q", item.Word, ch)
			}
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}