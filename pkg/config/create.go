package config

import (
	"bufio"
	"fmt"
	"github.com/inancgumus/screen"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

// File is the mtl-config.yml structure.
type File struct {
	CloudVision struct {
		CredentialsPath string `yaml:"credentialsPath"`
	} `yaml:"cloudVision"`
	Translation struct {
		SelectedService string `yaml:"selectedService"`
		Google          struct {
			APIKey string `yaml:"apiKey"`
		} `yaml:"google"`
		DeepL struct {
			APIKey  string `yaml:"apiKey"`
			Premium bool   `yaml:"premium"`
		} `yaml:"deepL"`
	} `yaml:"translation"`
}

// Create initiates an interactive setup for the config file.
// The "modify" parameter specifies if the setup is being used to modify
// an exiting config instead of staring from scratch.
func Create(modify bool) {
	var newConfig File
	var reader *bufio.Reader

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

	if newConfig.Translation.Google.APIKey == "" && newConfig.Translation.DeepL.APIKey == "" {
		defer Create(false)
		fmt.Println("You must enter at least 1 API key to use the application. Please try again.")
		return
	}

	// Set which service we will be using
	if newConfig.Translation.Google.APIKey == "" {
		newConfig.Translation.SelectedService = "deepL"
	} else if newConfig.Translation.DeepL.APIKey == "" {
		newConfig.Translation.SelectedService = "google"
	} else {
		if !modify || modifyConfirmation("Would you like to change which translation service you want to use?") {
			selectTLService(&newConfig)
		}
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
	fmt.Println("Input your Google Cloud Translation API key (leave blank if you don't have one):")
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

	// Check if DeepL Pro
	deepLPremium := false
	if deepLKey != "" {
		var deepLPremiumStr string
		for !(deepLPremiumStr == "yes" || deepLPremiumStr == "no") {
			fmt.Println("Is your DeepL API key for a DeepL Pro account? (yes/no):")
			reader = bufio.NewReader(os.Stdin)
			deepLPremiumStr, _ = reader.ReadString('\n')
			deepLPremiumStr = strings.TrimSuffix(deepLPremiumStr, "\r\n")
			deepLPremiumStr = strings.TrimSuffix(deepLPremiumStr, "\n")
			screen.Clear()
			screen.MoveTopLeft()
			log.Debugf("deepLPremiumStr: %v", deepLPremiumStr)
		}
		if deepLPremiumStr == "yes" {
			deepLPremium = true
		}
	}
	config.Translation.DeepL.Premium = deepLPremium
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

// SaveConfig saves the given ConfigFile object in "mtl-config.yml".
func SaveConfig(cfg File) {
	d, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile("mtl-config.yml", d, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
