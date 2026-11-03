package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// TranslateRequest represents the incoming translation request.
type TranslateRequest struct {
	Text       string `json:"text"`
	TargetLang string `json:"target_lang"` // "es" or "it"
}

// TranslateResponse represents the translation response.
type TranslateResponse struct {
	TranslatedText string           `json:"translated_text"`
	Vocabulary     []VocabularyItem `json:"vocabulary"`
	Cached         bool             `json:"cached"`
}

// VocabularyItem represents a single word/phrase entry.
type VocabularyItem struct {
	ID           string `json:"id,omitempty"`
	Word         string `json:"word"`
	Translation  string `json:"translation"`
	PartOfSpeech string `json:"part_of_speech,omitempty"`
	Difficulty   string `json:"difficulty,omitempty"`
	Example      string `json:"example,omitempty"`
}

// CacheKey generates a deterministic hash key for (text + targetLang).
// Exported for use by other handlers.
func CacheKey(text, targetLang string) string {
	h := sha256.New()
	h.Write([]byte(text + targetLang))
	return hex.EncodeToString(h.Sum(nil))
}

// TranslateHandler handles translation requests: checks cache, calls OpenRouter, extracts vocabulary.
func TranslateHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TranslateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.Text == "" || req.TargetLang == "" {
			http.Error(w, "text and target_lang are required", http.StatusBadRequest)
			return
		}

		key := CacheKey(req.Text, req.TargetLang)

		// Check cache
		cacheRecord, err := app.Dao().FindFirstRecordByExpr("translation_cache", fmt.Sprintf("source_hash = '%s'", key))
		if err == nil && cacheRecord != nil {
			// Update accessed_at
			cacheRecord.Set("accessed_at", time.Now().UTC().Format(time.RFC3339))
			app.Dao().SaveRecord(cacheRecord)

			// Get vocabulary from cache or extract
			vocabulary := extractVocabularyFromCache(app, req.Text, cacheRecord.GetString("translated_text"), req.TargetLang)

			resp := TranslateResponse{
				TranslatedText: cacheRecord.GetString("translated_text"),
				Vocabulary:     vocabulary,
				Cached:         true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Cache miss — call OpenRouter
		client := openai.NewClient(
			option.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")),
			option.WithBaseURL("https://openrouter.ai/api/v1"),
		)

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		// First, translate the text
		completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionCreateParams{
			Model: openai.String("openai/gpt-4o-mini"),
			Messages: []openai.ChatCompletionMessageParam{
				openai.ChatCompletionSystemMessageParam{
					Content: openai.String("You are a helpful translator. Translate the given text from English to the target language. Return ONLY the translation, nothing else."),
				},
				openai.ChatCompletionUserMessageParam{
					Content: openai.String(req.Text),
				},
			},
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("LLM error: %v", err), http.StatusInternalServerError)
			return
		}

		translatedText := ""
		if len(completion.Choices) > 0 {
			translatedText = completion.Choices[0].Message.Content
		}

		// Extract vocabulary using LLM
		vocabulary := extractVocabularyWithLLM(ctx, client, req.Text, translatedText, req.TargetLang)

		// Upsert vocabulary items
			vocabWithIDs := make([]VocabularyItem, len(vocabulary))
			for i, item := range vocabulary {
				id := upsertVocabulary(app, item, req.TargetLang)
				vocabWithIDs[i] = item
				vocabWithIDs[i].ID = id
			}

		// Store in cache
		cacheRecord = models.NewRecord()
		cacheRecord.Set("collection", "translation_cache")
		cacheRecord.Set("source_hash", key)
		cacheRecord.Set("source_text", req.Text)
		cacheRecord.Set("target_lang", req.TargetLang)
		cacheRecord.Set("translated_text", translatedText)
		cacheRecord.Set("tokens_used", len(req.Text)/4) // rough estimate
		cacheRecord.Set("created_at", time.Now().UTC().Format(time.RFC3339))
		cacheRecord.Set("accessed_at", time.Now().UTC().Format(time.RFC3339))

		if _, err := app.Dao().SaveRecord(cacheRecord); err != nil {
			log.Printf("Warning: failed to save cache: %v", err)
		}

		resp := TranslateResponse{
				TranslatedText: translatedText,
				Vocabulary:     vocabWithIDs,
				Cached:         false,
			}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// extractVocabularyWithLLM uses LLM to extract vocabulary with definitions, examples, POS.
func extractVocabularyWithLLM(ctx context.Context, client *openai.Client, source, translation, targetLang string) []VocabularyItem {
	langName := "Spanish"
	if targetLang == "it" {
		langName = "Italian"
	}

	prompt := fmt.Sprintf(`Extract key vocabulary from the following English text and its %s translation.
Return a JSON array of objects with these fields:
- word: the English word/phrase
- translation: the %s translation
- part_of_speech: noun, verb, adjective, adverb, etc.
- difficulty: beginner, intermediate, or advanced
- example: an example sentence in English using the word

Source text: %s
Translation: %s

Return ONLY the JSON array, no other text.`, langName, langName, source, translation)

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionCreateParams{
		Model: openai.String("openai/gpt-4o-mini"),
		Messages: []openai.ChatCompletionMessageParam{
			openai.ChatCompletionSystemMessageParam{
				Content: openai.String("You are a language learning assistant. Extract vocabulary as structured JSON."),
			},
			openai.ChatCompletionUserMessageParam{
				Content: openai.String(prompt),
			},
		},
		Temperature: openai.Float(0.3),
	})
	if err != nil {
		log.Printf("Vocabulary extraction failed: %v", err)
		return extractVocabularyFallback(source, translation, targetLang)
	}

	if len(completion.Choices) == 0 {
		return extractVocabularyFallback(source, translation, targetLang)
	}

	var items []VocabularyItem
	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	// Try to parse JSON array
	if err := json.Unmarshal([]byte(content), &items); err != nil {
		log.Printf("Failed to parse vocabulary JSON: %v, content: %s", err, content)
		return extractVocabularyFallback(source, translation, targetLang)
	}

	return items
}

// extractVocabularyFromCache retrieves vocabulary for cached translations.
func extractVocabularyFromCache(app *pocketbase.PocketBase, source, translation, targetLang string) []VocabularyItem {
	// Try to find existing vocabulary from the vocabulary table that matches words in source
	words := strings.Fields(source)
	seen := make(map[string]bool)
	var items []VocabularyItem

	for _, word := range words {
		clean := strings.Trim(word, ".,!?;:\"()[]{}")
		if len(clean) < 3 || seen[strings.ToLower(clean)] {
			continue
		}
		seen[strings.ToLower(clean)] = true

		// Look up in vocabulary table
		vocab, err := app.Dao().FindFirstRecordByExpr("vocabulary", fmt.Sprintf("word = '%s' AND language = '%s'", clean, targetLang))
		if err == nil && vocab != nil {
			items = append(items, VocabularyItem{
				ID:           vocab.Id,
				Word:         vocab.GetString("word"),
				Translation:  vocab.GetString("translation"),
				PartOfSpeech: vocab.GetString("part_of_speech"),
				Difficulty:   vocab.GetString("difficulty"),
				Example:      vocab.GetString("example_sentence"),
			})
		} else {
			items = append(items, VocabularyItem{
				Word:        clean,
				Translation: "",
				Difficulty:  "beginner",
			})
		}
	}

	return items
}

// extractVocabularyFallback is a simple fallback when LLM extraction fails.
func extractVocabularyFallback(source, translation, targetLang string) []VocabularyItem {
	// Simple word-based extraction (split on whitespace, filter short words)
	words := strings.Fields(source)
	var items []VocabularyItem
	seen := make(map[string]bool)

	for _, word := range words {
		// Clean word
		clean := strings.Trim(word, ".,!?;:\"()[]{}")
		if len(clean) < 3 || seen[strings.ToLower(clean)] {
			continue
		}
		seen[strings.ToLower(clean)] = true

		items = append(items, VocabularyItem{
			Word:        clean,
			Translation: "", // Would need alignment with translation
			Difficulty:  "beginner",
		})
	}

	return items
}

// upsertVocabulary creates or updates a vocabulary record and returns the record ID.
func upsertVocabulary(app *pocketbase.PocketBase, item VocabularyItem, lang string) string {
	// Check if word already exists
	existing, err := app.Dao().FindFirstRecordByExpr("vocabulary", fmt.Sprintf("word = '%s' AND language = '%s'", item.Word, lang))
	if err == nil && existing != nil {
		// Update
		existing.Set("translation", item.Translation)
		existing.Set("part_of_speech", item.PartOfSpeech)
		existing.Set("difficulty", item.Difficulty)
		existing.Set("example_sentence", item.Example)
		app.Dao().SaveRecord(existing)
		return existing.Id
	}

	// Create new
	record := models.NewRecord()
	record.Set("collection", "vocabulary")
	record.Set("word", item.Word)
	record.Set("translation", item.Translation)
	record.Set("language", lang)
	record.Set("part_of_speech", item.PartOfSpeech)
	record.Set("difficulty", item.Difficulty)
	record.Set("example_sentence", item.Example)
	app.Dao().SaveRecord(record)
	return record.Id
}

// EnrichHandler handles vocabulary enrichment (definitions, examples, difficulty).
func EnrichHandler(app *pocketbase.PocketBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
	}
}

// TranslateAndExtract is a shared function for translating text and extracting vocabulary.
// Used by lessons handler.
func TranslateAndExtract(app *pocketbase.PocketBase, text, targetLang string) (string, []VocabularyItem) {
	key := CacheKey(text, targetLang)

	// Check cache
	cacheRecord, err := app.Dao().FindFirstRecordByExpr("translation_cache", fmt.Sprintf("source_hash = '%s'", key))
	if err == nil && cacheRecord != nil {
		cacheRecord.Set("accessed_at", time.Now().UTC().Format(time.RFC3339))
		app.Dao().SaveRecord(cacheRecord)
		vocabulary := extractVocabularyFromCache(app, text, cacheRecord.GetString("translated_text"), targetLang)
		return cacheRecord.GetString("translated_text"), vocabulary
	}

	// Cache miss - call OpenRouter
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionCreateParams{
		Model: openai.String("openai/gpt-4o-mini"),
		Messages: []openai.ChatCompletionMessageParam{
			openai.ChatCompletionSystemMessageParam{
				Content: openai.String("You are a helpful translator. Translate the given text from English to the target language. Return ONLY the translation, nothing else."),
			},
			openai.ChatCompletionUserMessageParam{
				Content: openai.String(text),
			},
		},
	})
	if err != nil {
		log.Printf("Translation failed: %v", err)
		return "", []VocabularyItem{}
	}

	translatedText := ""
	if len(completion.Choices) > 0 {
		translatedText = completion.Choices[0].Message.Content
	}

	vocabulary := extractVocabularyWithLLM(ctx, client, text, translatedText, targetLang)

	// Upsert vocabulary items
	vocabWithIDs := make([]VocabularyItem, len(vocabulary))
	for i, item := range vocabulary {
		id := upsertVocabulary(app, item, targetLang)
		vocabWithIDs[i] = item
		vocabWithIDs[i].ID = id
	}

	// Store in cache
	cacheRecord = models.NewRecord()
	cacheRecord.Set("collection", "translation_cache")
	cacheRecord.Set("source_hash", key)
	cacheRecord.Set("source_text", text)
	cacheRecord.Set("target_lang", targetLang)
	cacheRecord.Set("translated_text", translatedText)
	cacheRecord.Set("tokens_used", len(text)/4)
	cacheRecord.Set("created_at", time.Now().UTC().Format(time.RFC3339))
	cacheRecord.Set("accessed_at", time.Now().UTC().Format(time.RFC3339))

	if _, err := app.Dao().SaveRecord(cacheRecord); err != nil {
		log.Printf("Warning: failed to save cache: %v", err)
	}

	return translatedText, vocabWithIDs
}