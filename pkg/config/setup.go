package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

// Setup loads the config at the given path into the given config object.
func Setup(configPath string, cfg *File) {
	log.Debugf("Config path: %v", configPath)
	viper.AddConfigPath(configPath)
	viper.SetConfigName("mtl-config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warning("Config file not found")
			return
		} else {
			log.Fatalf("Config file was found but another error was produced: %v", err)
		}
	}

	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("Unable to unmarshal config: %v", err)
	}
	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.CloudVision.CredentialsPath)
	if err != nil {
		log.Fatalf("Unable set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
	}
}

// Path returns the absolute path to the manga-translator files directory.
func Path() string {
	applicationPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	// Since applicationPath includes the application itself,
	// we must add ../ so that mtl-settings will be in the same directory as the application.
	settingsPath := filepath.Join(applicationPath, "../mtl")
	// Only create the directory if it does not exist.
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		err = os.Mkdir(settingsPath, os.ModePerm)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}

	return settingsPath
}
