package translate

import (
	"encoding/json"
	"errors"
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

// DeepLTranslate translates the given slice of strings from source language to target language using the DeepL API.
func DeepLTranslate(txt []string, source, target, apiKey string) ([]string, error) {
	log.WithFields(log.Fields{
		"sourceLanguage": source,
		"targetLanguage": target,
	}).Debug("Input languages")

	var baseUrl string
	if strings.HasSuffix(apiKey, ":fx") {
		baseUrl = "https://api-free.deepl.com/v2/"
	} else {
		baseUrl = "https://api.deepl.com/v2/"
	}

	params := url.Values{}
	for i := range txt {
		params.Add("text", txt[i])
	}
	if source != "" {
		params.Add("source_lang", source)
	}
	params.Add("auth_key", apiKey)
	params.Add("target_lang", target)
	reqBody := strings.NewReader(params.Encode())

	resp, err := http.Post(baseUrl+"translate", "application/x-www-form-urlencoded", reqBody)
	if err != nil {
		log.Errorf("http.Post: %v", err)
		return TranslationError("Translation request failed, ensure that your internet connection is stable.", txt), err
	}
	defer resp.Body.Close()
	log.Debugf("Translation request response: %v", resp)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body: %v \n", err)
		return TranslationError("Translation request failed, ensure that your internet connection is stable and your API key is correct.", txt), err
	}

	// Empty response body, something went wrong.
	if len(data) == 0 {
		log.Error("Empty response body from translation request")
		return TranslationError("Translation request failed, ensure that your API key is correct.", txt), errors.New("empty response body from translation request")
	}

	var jsonData DeepLResponse
	if err = json.Unmarshal(data, &jsonData); err != nil {
		log.Errorf("Parse response failed: %v", err)
		return TranslationError("Translation request failed, ensure that your internet connection is stable and your API key is correct.", txt), err
	}
	log.Debugf("Translation response body: %v", jsonData)

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		log.Error("Non-200 status code")
		return TranslationError("Translation request failed, ensure that your API key and source/target languages are correct.", txt), errors.New("non-200 status code")
	}

	var translated []string
	for _, t := range jsonData.Translations {
		translated = append(translated, t.Text)
	}

	log.WithField("text", translated).Info("Translated Text")

	return translated, nil
}
