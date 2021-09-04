package translate

import (
	"encoding/json"
	"github.com/cameronkinsella/manga-translator/pkg/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type DeepLResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
	Message string `json:"message"`
}

// TranslationError creates a slice of strings all containing the given message.
// Used to put an error message in the "Translated Text" section of the GUI.
func TranslationError(message string, txt []string) []string {
	var failureMessage []string
	for range txt {
		failureMessage = append(
			failureMessage,
			message,
		)
	}
	return failureMessage
}

// DeepLTranslate translates the given slice of strings from Japanese to English using the DeepL API.
func DeepLTranslate(txt []string, cfg config.File) []string {
	var reqURL string
	if cfg.Translation.DeepL.Premium {
		reqURL = "https://api.deepl.com/v2/translate"
	} else {
		reqURL = "https://api-free.deepl.com/v2/translate"
	}

	params := url.Values{}
	for i := range txt {
		params.Add("text", txt[i])
	}
	params.Add("auth_key", cfg.Translation.DeepL.APIKey)
	params.Add("source_lang", `JA`)
	params.Add("target_lang", `EN`)
	reqBody := strings.NewReader(params.Encode())

	resp, err := http.Post(reqURL, "application/x-www-form-urlencoded", reqBody)
	if err != nil {
		log.Errorf("http.Post: %v", err)
		return TranslationError("Translation request failed, ensure that your internet connection is stable.", txt)
	}
	defer resp.Body.Close()
	log.Debugf("Translation request response: %v", resp)

	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		log.Errorf("Error reading response body: %v \n", err2)
		return TranslationError("Translation request failed, ensure that your internet connection is stable and your API key is correct.", txt)
	}

	// Empty response body, something went wrong.
	if len(data) == 0 {
		log.Error("Empty response body from translation request")
		return TranslationError("Translation request failed, ensure that your API key is correct.", txt)
	}

	var jsonData DeepLResponse
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Errorf("Parse response failed: %v", err)
		return TranslationError("Translation request failed, ensure that your internet connection is stable and your API key is correct.", txt)
	}
	log.Debugf("Translation response body: %v", jsonData)

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		log.Error("Non-200 status code")
		if strings.HasPrefix(jsonData.Message, "Wrong endpoint.") {
			// Wrong tier in config, fix the config and retry.
			log.Warning("Wrong DeepL translation tier selected. Fixing config file.")
			cfg.Translation.DeepL.Premium = !cfg.Translation.DeepL.Premium
			config.SaveConfig(cfg)
			return DeepLTranslate(txt, cfg)
		}

		return TranslationError("Translation request failed, ensure that your API key is correct.", txt)
	}

	var translated []string
	for _, t := range jsonData.Translations {
		translated = append(translated, t.Text)
	}

	log.WithField("text", translated).Info("Translated Text")

	return translated
}
