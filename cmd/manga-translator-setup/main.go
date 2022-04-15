package main

import (
	"github.com/cameronkinsella/manga-translator/pkg/config"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func main() {
	// Set up logging.
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	settings := config.Path()
	logPath := filepath.Join(settings, "mtl-logrus.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		log.SetOutput(f)
	} else {
		log.Warning("Failed to log to file, using default stderr")
	}
	defer f.Close()

	// Try loading config file.
	var cfg config.File
	config.Setup(settings, &cfg)

	// We only want to start from scratch if there is no existing config, otherwise we modify existing config.
	modify := cfg != config.File{}

	config.Create(modify)
}
