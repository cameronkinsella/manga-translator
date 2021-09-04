package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

// Setup loads the config at the given path into the given config object.
func Setup(configPath string, cfg *File) {
	log.Debugf("Config path: %v", configPath)
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warning("Config file not found")
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
