package translate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
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

// DeepLTranslate translates the given slice of strings from source language to target language using the DeepL API.
func DeepLTranslate(txt []string, source, target, apiKey string) ([]string, error) {
	log.WithFields(log.Fields{
		"sourceLanguage": source,
		"targetLanguage": target,
	}).Debug("Input languages")

	// Determine base URL
	baseURL := "https://api.deepl.com/v2/"
	if strings.HasSuffix(apiKey, ":fx") {
		baseURL = "https://api-free.deepl.com/v2/"
	}

	// Build form parameters
	params := url.Values{}
	for _, t := range txt {
		params.Add("text", t)
	}
	if source != "" {
		params.Add("source_lang", source)
	}
	if target == "" {
		target = "EN-US"
	}
	params.Add("target_lang", target)
	params.Add("model_type", "quality_optimized")

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"translate",
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return TranslationError("Failed to create translation request.", txt), err
	}

	// Required headers
	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("client.Do: %v", err)
		return TranslationError("Translation request failed, check your internet connection.", txt), err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("ReadAll: %v", err)
		return TranslationError("Failed to read translation response.", txt), err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Errorf("DeepL API error (%d): %s", resp.StatusCode, string(data))
		return TranslationError("Translation request failed, ensure your API key and languages are correct.", txt),
			fmt.Errorf("deepl api error: %s", resp.Status)
	}

	if len(data) == 0 {
		return TranslationError("Empty response from translation service.", txt),
			errors.New("empty response body")
	}

	var jsonData DeepLResponse
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Errorf("Unmarshal failed: %v", err)
		return TranslationError("Failed to parse translation response.", txt), err
	}

	var translated []string
	for _, t := range jsonData.Translations {
		translated = append(translated, t.Text)
	}

	log.WithField("text", translated).Info("Translated Text")
	return translated, nil
}
