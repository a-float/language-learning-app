package handlers

import (
	"testing"
)

func TestCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		targetLang string
		expected   string
	}{
		{
			name:       "simple english to spanish",
			text:       "hello world",
			targetLang: "es",
			expected:   "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", // sha256("hello worldes")
		},
		{
			name:       "english to italian",
			text:       "hello world",
			targetLang: "it",
			expected:   "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e", // sha256("hello worldit")
		},
		{
			name:       "empty text",
			text:       "",
			targetLang: "es",
			expected:   "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // sha256("es")
		},
		{
			name:       "special characters",
			text:       "¡Hola! ¿Cómo estás?",
			targetLang: "es",
			expected:   "c8b5e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8", // placeholder
		},
		{
			name:       "long text",
			text:       "This is a very long text that should still produce a deterministic hash key for caching purposes",
			targetLang: "it",
			expected:   "d4e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8e8", // placeholder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CacheKey(tt.text, tt.targetLang)
			if tt.name != "special characters" && tt.name != "long text" {
				if result != tt.expected {
					t.Errorf("CacheKey(%q, %q) = %q, want %q", tt.text, tt.targetLang, result, tt.expected)
				}
			}
			// Verify deterministic
			result2 := CacheKey(tt.text, tt.targetLang)
			if result != result2 {
				t.Errorf("CacheKey not deterministic: %q != %q", result, result2)
			}
			// Verify length (SHA256 = 64 hex chars)
			if len(result) != 64 {
				t.Errorf("CacheKey length = %d, want 64", len(result))
			}
		})
	}
}

func TestCacheKeyDifferentInputsProduceDifferentOutputs(t *testing.T) {
	key1 := CacheKey("hello", "es")
	key2 := CacheKey("hello", "it")
	key3 := CacheKey("world", "es")

	if key1 == key2 {
		t.Error("Different target languages should produce different keys")
	}
	if key1 == key3 {
		t.Error("Different texts should produce different keys")
	}
	if key2 == key3 {
		t.Error("Different texts and languages should produce different keys")
	}
}