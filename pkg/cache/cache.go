package cache

import (
	"encoding/gob"
	"errors"
	"github.com/cameronkinsella/manga-translator/pkg/detect"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type Data struct {
	Hash    string
	Service string
	Blocks  []detect.TextBlock
}

// read creates a cache if one doesn't already exist, reads the data from the cache,
// and returns it as a slice of CacheData.
func read() []Data {
	var cacheData []Data

	cachePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	cachePath = filepath.Join(cachePath, "mtl-cache.bin")
	cacheFile, err := os.Open(cachePath)
	if errors.Is(err, os.ErrNotExist) {
		// handle the case where the file doesn't exist
		cacheFile, err = os.Create(cachePath)
		if err != nil {
			log.Fatal(err)
		}
		defer cacheFile.Close()

		enc := gob.NewEncoder(cacheFile)
		if err := enc.Encode(cacheData); err != nil {
			log.Fatalf("Cache creation failed: %v", err)
		}

		return cacheData
	} else if err != nil {
		log.Fatal(err)
	}
	defer cacheFile.Close()

	dec := gob.NewDecoder(cacheFile)
	if err := dec.Decode(&cacheData); err != nil {
		log.Fatal(err)
	}
	return cacheData
}

// Check returns the text blocks of the given image hash if it is in cache, otherwise returns nil.
func Check(h string, service string) []detect.TextBlock {
	cacheData := read()

	for _, data := range cacheData {
		if h == data.Hash && data.Service == service {
			log.Info("Image found in cache, skipping API requests.")
			return data.Blocks
		}
	}

	log.Info("Image not found in cache, performing API requests.")
	return nil
}

// Add adds a new entry to the cache.
func Add(h string, service string, blocks []detect.TextBlock) {
	log.Debugf("Adding new image to cache. sha256:%v", h)
	cacheData := read()

	cachePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	cachePath = filepath.Join(cachePath, "mtl-cache.bin")
	cacheFile, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer cacheFile.Close()

	newData := Data{
		Hash:    h,
		Service: service,
		Blocks:  blocks,
	}
	cacheData = append(cacheData, newData)
	enc := gob.NewEncoder(cacheFile)
	if err := enc.Encode(cacheData); err != nil {
		log.Fatalf("Cache write failed: %v", err)
	}
}
