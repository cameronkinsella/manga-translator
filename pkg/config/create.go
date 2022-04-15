package config

import (
	"bufio"
	"cloud.google.com/go/translate"
	"context"
	"encoding/json"
	"fmt"
	"github.com/inancgumus/screen"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// File is the mtl/mtl-config.yml structure.
type File struct {
	CloudVision struct {
		CredentialsPath string `yaml:"credentialsPath"`
	} `yaml:"cloudVision"`
	Translation struct {
		SelectedService string `yaml:"selectedService"`
		SourceLanguage  string `yaml:"sourceLanguage,omitempty"`
		TargetLanguage  string `yaml:"targetLanguage"`
		Google          struct {
			APIKey string `yaml:"apiKey,omitempty"`
		} `yaml:"google,omitempty"`
		DeepL struct {
			APIKey string `yaml:"apiKey,omitempty"`
		} `yaml:"deepL,omitempty"`
	} `yaml:"translation"`
}

// deepLLanguage is the structure of language objects returned from the language list API.
type deepLLanguage struct {
	Language string `json:"language"`
	Name     string `json:"name"`
}

// languageObj is used to map ISO-639-1 codes to their respective languages.
type languageObj struct {
	Code     string
	Language string
}

// Create initiates an interactive setup for the config file.
// The "modify" parameter specifies if the setup is being used to modify
// an exiting config instead of staring from scratch.
func Create(modify bool) {
	var newConfig File
	var reader *bufio.Reader
	screen.Clear()
	screen.MoveTopLeft()

	if modify {
		var startOver string
		for !(startOver == "1" || startOver == "2") {
			fmt.Println(
				"You already have a config configured, would you like to start from scratch or modify the existing config? (type 1 or 2):\n" +
					"[1] Start from scratch\n" +
					"[2] Modify existing",
			)
			reader = bufio.NewReader(os.Stdin)
			startOver, _ = reader.ReadString('\n')
			startOver = strings.TrimSuffix(startOver, "\r\n")
			startOver = strings.TrimSuffix(startOver, "\n")
			screen.Clear()
			screen.MoveTopLeft()
		}
		if startOver == "1" {
			defer Create(false)
			return
		}
		err := viper.Unmarshal(&newConfig)
		if err != nil {
			log.Fatalf("Unable to unmarshal existing config: %v", err)
		}
	}

	// Google Cloud Vision API Key
	if !modify || modifyConfirmation("Would you like to change your Google Cloud Vision API Key?") {
		setupVisionAPIKey(&newConfig)
	}

	// Google Cloud Translation API Key
	if !modify || modifyConfirmation("Would you like to change your Google Cloud Translation API Key?") {
		setupGoogleAPIKey(&newConfig)
	}

	// DeepL Translation API Key
	if !modify || modifyConfirmation("Would you like to change your DeepL Translation configuration?") {
		setupDeepLConfig(&newConfig)
	}

	updateLang := false

	// Set which service we will be using
	if newConfig.Translation.DeepL.APIKey == "" {
		// Only Google key or no API keys
		newConfig.Translation.SelectedService = "google"
	} else if newConfig.Translation.Google.APIKey == "" {
		// Only DeepL key
		newConfig.Translation.SelectedService = "deepL"
	} else {
		// Both Google key and DeepL key. Must choose which one to use.
		if !modify || modifyConfirmation("Would you like to change which translation service you want to use?") {
			prevService := newConfig.Translation.SelectedService
			selectTLService(&newConfig)
			if modify && prevService != newConfig.Translation.SelectedService {
				log.WithFields(log.Fields{
					"prevService": prevService,
					"newService":  newConfig.Translation.SelectedService,
				}).Debug("Translation service changed")

				fmt.Println("Successfully changed the translation service.\n" +
					`You will need to update your "source language" and "target language".`)
				fmt.Println("Press 'Enter' to continue.")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
				screen.Clear()
				screen.MoveTopLeft()
				updateLang = true
			}
		}
	}

	// Source language
	if !modify || updateLang || modifyConfirmation("Would you like to change your source language?") {
		setupSourceLanguage(&newConfig)
	}

	// Target language
	if !modify || updateLang || modifyConfirmation("Would you like to change your target language?") {
		setupTargetLanguage(&newConfig)
	}

	SaveConfig(newConfig)
	fmt.Println(`Config setup complete! Run the "manga-translator-setup" application again if you want to modify it.`)
	fmt.Println("Press 'Enter' to exit.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(0)
}

// setupVisionAPIKey initiates an interactive prompt to set the Cloud Vision API key for the given config.
func setupVisionAPIKey(config *File) {
	var credentialsPath string
	for credentialsPath == "" {
		fmt.Println("Input the path to the service account credentials for the Vision API (required):")
		reader := bufio.NewReader(os.Stdin)
		credentialsPath, _ = reader.ReadString('\n')
		credentialsPath = strings.TrimSuffix(credentialsPath, "\r\n")
		credentialsPath = strings.TrimSuffix(credentialsPath, "\n")
		screen.Clear()
		screen.MoveTopLeft()
		log.Debugf("credentialsPath: %v", credentialsPath)
	}
	config.CloudVision.CredentialsPath = credentialsPath
}

// setupGoogleAPIKey initiates an interactive prompt to set the Google Translation API key for the given config.
func setupGoogleAPIKey(config *File) {
	fmt.Println("Input your Google Cloud Translation API key (leave blank if you don't have one or want to use your service account key instead):")
	reader := bufio.NewReader(os.Stdin)
	googleTranslateKey, _ := reader.ReadString('\n')
	googleTranslateKey = strings.TrimSuffix(googleTranslateKey, "\r\n")
	googleTranslateKey = strings.TrimSuffix(googleTranslateKey, "\n")
	config.Translation.Google.APIKey = googleTranslateKey
	screen.Clear()
	screen.MoveTopLeft()
	log.Debugf("googleTranslateKey: %v", googleTranslateKey)
}

// setupDeepLConfig initiates an interactive prompt to set the DeepL API key for the given config.
func setupDeepLConfig(config *File) {
	fmt.Println("Input your DeepL API key (leave blank if you don't have one):")
	reader := bufio.NewReader(os.Stdin)
	deepLKey, _ := reader.ReadString('\n')
	deepLKey = strings.TrimSuffix(deepLKey, "\r\n")
	deepLKey = strings.TrimSuffix(deepLKey, "\n")
	config.Translation.DeepL.APIKey = deepLKey
	screen.Clear()
	screen.MoveTopLeft()
	log.Debugf("deepLKey: %v", deepLKey)
}

// setupDeepLConfig initiates an interactive prompt to set the desired translation service for the given config.
func selectTLService(config *File) {
	var selectedService string
	for !(selectedService == "1" || selectedService == "2") {
		fmt.Println(
			"You have configured both Google Cloud Translation and DeepL, which would you like to use? (type 1 or 2):\n" +
				"[1] Google Cloud Translation\n" +
				"[2] DeepL Translation",
		)
		reader := bufio.NewReader(os.Stdin)
		selectedService, _ = reader.ReadString('\n')
		selectedService = strings.TrimSuffix(selectedService, "\r\n")
		selectedService = strings.TrimSuffix(selectedService, "\n")
		screen.Clear()
		screen.MoveTopLeft()
		log.Debugf("selectedService: %v", selectedService)
	}
	if selectedService == "1" {
		selectedService = "google"
	} else {
		selectedService = "deepL"
	}
	config.Translation.SelectedService = selectedService
}

// setupSourceLanguage initiates an interactive prompt to set the desired source language for the given config.
func setupSourceLanguage(config *File) {
	supportedLangs := getSupportedLanguages(config, "source")
	var sourceLang string
	for !isSupportedLanguage(supportedLangs, sourceLang) {
		fmt.Println("Enter 'list' to display all source languages (the language you will translate from).\n" +
			"Input the source language ISO-639-1 code (leave blank to automatically detect language):")
		reader := bufio.NewReader(os.Stdin)
		sourceLang, _ = reader.ReadString('\n')
		sourceLang = strings.TrimSuffix(sourceLang, "\r\n")
		sourceLang = strings.TrimSuffix(sourceLang, "\n")
		if sourceLang == "list" {
			screen.Clear()
			screen.MoveTopLeft()
			fmt.Println("\"code\": Language\n——————————————————")
			for _, i := range supportedLangs {
				fmt.Printf("%q: %s\n", i.Code, i.Language)
			}
			fmt.Println("")
			defer setupSourceLanguage(config)
			return
		}
		screen.Clear()
		screen.MoveTopLeft()
		log.Debugf("sourceLanguage: %v", sourceLang)

		// Blank source will delete it from config. Translators will then choose automatic detection.
		if sourceLang == "" {
			break
		}
	}
	config.Translation.SourceLanguage = sourceLang
}

// setupTargetLanguage initiates an interactive prompt to set the desired target language for the given config.
func setupTargetLanguage(config *File) {
	supportedLangs := getSupportedLanguages(config, "target")
	var targetLang string
	for !isSupportedLanguage(supportedLangs, targetLang) {
		fmt.Println("Enter 'list' to display all target languages (the language you will translate to).\n" +
			"Input the target language ISO-639-1 code (leave blank for english):")
		reader := bufio.NewReader(os.Stdin)
		targetLang, _ = reader.ReadString('\n')
		targetLang = strings.TrimSuffix(targetLang, "\r\n")
		targetLang = strings.TrimSuffix(targetLang, "\n")
		if targetLang == "list" {
			screen.Clear()
			screen.MoveTopLeft()
			fmt.Println("\"code\": Language\n——————————————————")
			for _, i := range supportedLangs {
				fmt.Printf("%q: %s\n", i.Code, i.Language)
			}
			fmt.Println("")
			defer setupTargetLanguage(config)
			return
		} else if targetLang == "" {
			if config.Translation.SelectedService == "google" {
				targetLang = "en"
			} else {
				targetLang = "EN-US"
			}
		}
		screen.Clear()
		screen.MoveTopLeft()
		log.Debugf("targetLanguage: %v", targetLang)
	}
	config.Translation.TargetLanguage = targetLang
}

// isSupportedLanguage returns if the given ISO-639-1 language code is contained in the given slice of languages.
func isSupportedLanguage(languageList []languageObj, languageCode string) bool {
	for _, i := range languageList {
		if i.Code == languageCode {
			return true
		}
	}
	return false
}

// getSupportedLanguages returns a list of languages which are supported for the given language type (source or target)
// using the translation service written in the Translation.SelectedService field of the given config.
func getSupportedLanguages(config *File, languageType string) (languageList []languageObj) {
	if config.Translation.SelectedService == "google" {
		ctx := context.Background()

		// Display results in english.
		lang, err := language.Parse("en")
		if err != nil {
			log.Errorf("language.Parse: %v", err)
			fmt.Printf("Error: language.Parse: %v", err)
			return
		}

		var client *translate.Client
		if config.Translation.Google.APIKey == "" {
			err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.CloudVision.CredentialsPath)
			if err != nil {
				log.Fatalf("Unable set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
			}
			client, err = translate.NewClient(ctx)
			if err != nil {
				log.Fatalf("NewClient: %v", err)
			}
		} else {
			apiKeyOption := option.WithAPIKey(config.Translation.Google.APIKey)
			client, err = translate.NewClient(ctx, apiKeyOption)
			if err != nil {
				log.Fatalf("translate.NewClient: %v", err)
			}
		}
		defer client.Close()

		langs, err := client.SupportedLanguages(ctx, lang)
		if err != nil {
			log.Errorf("SupportedLanguages: %v", err)
			fmt.Printf("Error: SupportedLanguages: %v", err)
			return
		}

		for _, lang := range langs {
			languageList = append(languageList, languageObj{lang.Tag.String(), lang.Name})
		}
	} else {
		var baseUrl string
		if strings.HasSuffix(config.Translation.DeepL.APIKey, ":fx") {
			baseUrl = "https://api-free.deepl.com/v2/"
		} else {
			baseUrl = "https://api.deepl.com/v2/"
		}

		params := url.Values{}
		params.Add("auth_key", config.Translation.DeepL.APIKey)
		params.Add("type", languageType)

		reqBody := strings.NewReader(params.Encode())

		resp, err := http.Post(baseUrl+"languages", "application/x-www-form-urlencoded", reqBody)
		if err != nil {
			log.Errorf("http.Post: %v", err)
			return
		}
		defer resp.Body.Close()
		log.Debugf("Language list response: %v", resp)

		data, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			log.Errorf("Error reading response body: %v \n", err2)
			return
		}

		// Empty response body, something went wrong.
		if len(data) == 0 {
			log.Error("Empty response body from language list request")
			return
		}

		var jsonData []deepLLanguage
		if err := json.Unmarshal(data, &jsonData); err != nil {
			log.Errorf("Parse response failed: %v", err)
			return
		}

		for _, i := range jsonData {
			languageList = append(languageList, languageObj{i.Language, i.Name})
		}
	}
	return
}

// modifyConfirmation initiates an interactive confirmation prompt and returns the response as a bool.
func modifyConfirmation(msg string) bool {
	var response string

	for !(response == "yes" || response == "no") {
		fmt.Println(msg + "(yes/no):")
		reader := bufio.NewReader(os.Stdin)
		response, _ = reader.ReadString('\n')
		response = strings.TrimSuffix(response, "\r\n")
		response = strings.TrimSuffix(response, "\n")
		screen.Clear()
		screen.MoveTopLeft()
		log.Debugf("confirmation response: %v", response)
	}
	if response == "yes" {
		return true
	}
	return false
}

// SaveConfig saves the given ConfigFile object in "mtl/mtl-config.yml".
func SaveConfig(cfg File) {
	d, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	configPath := filepath.Join(Path(), "mtl-config.yml")
	err = ioutil.WriteFile(configPath, d, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
