package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

type Word struct {
	Word        string `json:"phrase" jsonschema_description:"Word worth learning in the source text. Must be one word, but for verbs may include the person."`
	Lemma       string `json:"lemma" jsonschema_description:"Lemma of the word - for verbs like went, gone, goes it would be just go"`
	Translation string `json:"translation" jsonschema_description:"Translation of the word."`
}

type Phrase struct {
	Phrase      string `json:"phrase" jsonschema_description:"Phrase found in the source text. Must be longer that one word."`
	Translation string `json:"translation" jsonschema_description:"Short translation of the phrase."`
}

type LanguageLearningMaterial struct {
	Name                   string   `json:"name" jsonschema_description:"A short name summarizing the soruce text"`
	DetectedSourceLanguage string   `json:"detectedSourceLanguage" jsonschema_description:"The detected source language"`
	Phrases                []Phrase `json:"phrases" jsonschema_description:"Phrases of significance found in the source text"`
	Words                  []Word   `json:"words" jsonschema_description:"Words of significance found in the source text"`
}

// Structured Outputs uses a subset of JSON schema
// These flags are necessary to comply with the subset
func GenerateSchema[T any]() (map[string]any, error) {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)

	data, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	// Preserve schema values as raw JSON so integer constraints are not rounded
	// through float64 before the SDK serializes the request.
	var rawSchema map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return nil, err
	}
	result := make(map[string]any, len(rawSchema))
	for key, value := range rawSchema {
		result[key] = value
	}
	return result, nil
}

func saveLearningMaterial(app *pocketbase.PocketBase, text string, material LanguageLearningMaterial) error {
	app.Logger().Info("Saving material")
	err := app.RunInTransaction(func(txApp core.App) error {
		materialsCollection, err := txApp.FindCollectionByNameOrId("learningMaterials")
		if err != nil {
			return err
		}

		wordsCollection, err := txApp.FindCollectionByNameOrId("words")
		if err != nil {
			return err
		}

		phrasesCollection, err := txApp.FindCollectionByNameOrId("phrases")
		if err != nil {
			return err
		}

		wordIDs := []string{}
		phraseIDs := []string{}

		for _, word := range material.Words {
			record := core.NewRecord(wordsCollection)
			record.Set("word", word.Word)
			record.Set("lemma", word.Lemma)
			record.Set("translation", word.Translation)

			if err := txApp.Save(record); err != nil {
				return fmt.Errorf("Failed to save word: %w", err)
			}
			wordIDs = append(wordIDs, record.Id)
		}

		for _, phrase := range material.Phrases {
			record := core.NewRecord(phrasesCollection)
			record.Set("phrase", phrase.Phrase)
			record.Set("translation", phrase.Translation)

			if err := txApp.Save(record); err != nil {
				return fmt.Errorf("Failed to save phrase: %w", err)
			}
			phraseIDs = append(phraseIDs, record.Id)
		}

		record := core.NewRecord(materialsCollection)
		record.Set("text", text)
		record.Set("sourceLanguage", material.DetectedSourceLanguage)
		record.Set("targetLanguage", "English")
		record.Set("words", wordIDs)
		record.Set("phrases", phraseIDs)

		if err := txApp.Save(record); err != nil {
			return fmt.Errorf("Failed to save learning material: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	app.Logger().Info(fmt.Sprintf("Created learning material with %d words and %d phrases", len(material.Words), len(material.Phrases)))
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := pocketbase.New()
	client := openai.NewClient(
		option.WithBaseURL("https://openrouter.ai/api/v1"),
		option.WithAPIKey(os.Getenv("OPEN_ROUTER_KEY")),
	)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/static/{path...}", apis.Static(os.DirFS("./pb_public"), false))
		se.Router.GET("/favicon.ico", func(e *core.RequestEvent) error {
			http.ServeFile(e.Response, e.Request, "./pb_public/favicon.ico")
			return nil
		})

		se.Router.GET("/api/hello/{name}", func(e *core.RequestEvent) error {
			name := e.Request.PathValue("name")
			return e.String(http.StatusOK, "Hello "+name)
		})

		se.Router.POST("/api/answer", func(e *core.RequestEvent) error {
			data := struct {
				Text           string `json:"text"`
				TargetLanguage string `json:"targetLanguage"`
			}{}

			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to parse request JSON", err) //
			}

			app.Logger().Debug("Received translate request")

			schema, err := GenerateSchema[LanguageLearningMaterial]()
			if err != nil {
				panic(err)
			}

			input := fmt.Sprintf(`
				Target language: %s

				Text:
				%s
				`, data.TargetLanguage, data.Text)

			resp, err := client.Responses.New(e.Request.Context(), responses.ResponseNewParams{
				Instructions: openai.String("You are a professional translator and teacher. " +
					"Analyze the input text. Detect what languages its from" +
					"Prepare learning materials based on that text to allow the learner to undersand it" +
					"Detect the source language of the provided text" +
					"Return translation and meanings for the essentials words." +
					"Return phrases of importance that are not translateable word for word and explain their meaning"),
				Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(input)},
				Model: "nex-agi/nex-n2.5-mini:free",
				Text: responses.ResponseTextConfigParam{
					Format: responses.ResponseFormatTextConfigUnionParam{
						OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
							Name:        "language_learning_material",
							Description: openai.String("Learning materials based on the text"),
							Schema:      schema,
							Strict:      openai.Bool(true),
						},
					},
				},
			})

			app.Logger().Debug("AI Response", "model", resp.Model, "responseLength", len(resp.OutputText()))

			if err != nil {
				panic(err)
			}

			var material LanguageLearningMaterial

			if err := json.Unmarshal([]byte(resp.OutputText()), &material); err != nil {
				return fmt.Errorf("failed to parse learning material: %w", err)
			}

			err = saveLearningMaterial(app, data.Text, material)

			if err != nil {
				panic(err)
			}

			return e.JSON(http.StatusOK, map[string]any{"answer": resp.OutputText()})
		})

		return se.Next()
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: false,
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
