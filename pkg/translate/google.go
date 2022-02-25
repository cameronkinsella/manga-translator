package translate

import (
	"cloud.google.com/go/translate"
	"context"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// GoogleTranslate translates the given slice of strings from source language to target language using the Google Cloud Translation API.
func GoogleTranslate(txt []string, source, target, apiKey string) ([]string, error) {
	log.WithFields(log.Fields{
		"sourceLanguage": source,
		"targetLanguage": target,
	}).Debug("Input languages")

	options := translate.Options{
		Format: translate.Text,
	}

	// Set source language if one was given. Otherwise, do not specify (automatically detect source language).
	if source != "" {
		sourceLang, err := language.Parse(source)
		if err != nil {
			log.Errorf("language.Parse: %v", err)
			return TranslationError("Invalid source language selected in config.", txt), err
		}
		options.Source = sourceLang
	}

	// Support configs which do not have "targetLanguage" (version <=1.2.0)
	if target == "" {
		target = "en"
	}

	targetLang, err := language.Parse(target)
	if err != nil {
		log.Errorf("language.Parse: %v", err)
		return TranslationError("Invalid target language selected in config.", txt), err
	}

	ctx := context.Background()
	var client *translate.Client
	if apiKey == "" {
		client, err = translate.NewClient(ctx)
		if err != nil {
			log.Errorf("NewClient: %v", err)
			return TranslationError("Translation request failed, ensure that the absolute path given for your Vision API service account key is correct", txt), err
		}
	} else {
		apiKeyOption := option.WithAPIKey(apiKey)
		client, err = translate.NewClient(ctx, apiKeyOption)
		if err != nil {
			log.Errorf("NewClient: %v", err)
			return TranslationError("Translation request failed, ensure that your API key is correct.", txt), err
		}
	}
	defer client.Close()

	resp, err := client.Translate(ctx, txt, targetLang, &options)
	log.Debug(resp)

	if err != nil {
		log.Errorf("Translate: %v", err)
		return TranslationError("Translation request failed, ensure that your API key is correct.", txt), err
	}

	var translated []string
	for _, t := range resp {
		translated = append(translated, t.Text)
	}

	return translated, nil
}
