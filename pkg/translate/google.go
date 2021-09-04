package translate

import (
	"cloud.google.com/go/translate"
	"context"
	"github.com/cameronkinsella/manga-translator/pkg/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// GoogleTranslate translates the given slice of strings from Japanese to English using the Google Cloud Translation API.
func GoogleTranslate(txt []string, cfg config.File) []string {
	ctx := context.Background()
	apiKeyOption := option.WithAPIKey(cfg.Translation.Google.APIKey)
	client, err := translate.NewClient(ctx, apiKeyOption)
	if err != nil {
		log.Errorf("NewClient: %v", err)
		return TranslationError("Translation request failed, ensure that your API key is correct.", txt)
	}
	defer client.Close()

	resp, err := client.Translate(ctx, txt,
		language.English,
		&translate.Options{
			Source: language.Japanese,
			Format: translate.Text,
		},
	)
	log.Debug(resp)

	if err != nil {
		log.Errorf("Translate: %v", err)
		return TranslationError("Translation request failed, ensure that your API key is correct.", txt)
	}

	var translated []string
	for _, t := range resp {
		translated = append(translated, t.Text)
	}

	return translated
}
