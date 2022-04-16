package main

import (
	"encoding/hex"
	"flag"
	"gioui.org/app"
	"gioui.org/unit"
	"github.com/cameronkinsella/manga-translator/pkg/config"
	imageW "github.com/cameronkinsella/manga-translator/pkg/image"
	"github.com/cameronkinsella/manga-translator/pkg/window"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var maxDim float32 = 1000 // hard-coded

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

	// Parse flags.
	urlImagePtr := flag.Bool("url", false, "Use an image from a URL instead of a local file.")
	clipImagePtr := flag.Bool("clip", false, "Use an image from the clipboard.") // overrides url
	flag.Parse()
	log.Infof("Use URL image: %v", *urlImagePtr)
	log.Infof("Use clipboard image: %v", *clipImagePtr)

	// Set up config, create new config if necessary.
	var cfg config.File
	config.Setup(settings, &cfg)

	// Open/download selected image and get its info.
	if len(flag.Args()) == 0 && !*clipImagePtr {
		log.Fatal("No path or URL given.")
	}
	var imgPath []string
	if !*clipImagePtr {
		imgPath = flag.Args()
		log.Infof("All Selected Image(s): %v", imgPath)
	}

	var img []window.ImageInfo

	for _, imgPath := range imgPath {
		log.Debugf("Getting image info for: %v", imgPath)
		imgFile, imgHash := imageW.Open(imgPath, *urlImagePtr, *clipImagePtr)
		hashInBytes := imgHash.Sum(nil)
		imgHashStr := hex.EncodeToString(hashInBytes)
		log.Debugf("Hash: %v", imgHashStr)
		dims := imageW.GetDimensions(imgFile)
		log.Debugf("Image Dimensions: %v", dims)

		newImage := window.ImageInfo{
			Image:      imgFile,
			Path:       imgPath,
			Hash:       imgHashStr,
			Dimensions: dims,
		}
		img = append(img, newImage)
	}

	options := window.Options{
		Url:  *urlImagePtr,
		Clip: *clipImagePtr,
	}

	// We need this ratio to scale the image down/up to the required starting size.
	ratio := imageW.GetRatio(img[0].Dimensions, maxDim)

	go func() {
		// Create new window.
		w := app.NewWindow(
			app.Title("Manga Translator"),
			app.Size(unit.Dp(ratio*img[0].Dimensions.Width), unit.Dp(ratio*img[0].Dimensions.Height)),
			app.MinSize(unit.Dp(600), unit.Dp(300)),
		)

		if err := window.DrawFrame(w, img, options, cfg); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
