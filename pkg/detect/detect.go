package detect

import (
	"bytes"
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"image/color"
	_ "image/jpeg"
	"os"
	"strings"
)

var borderColors = []color.NRGBA{
	{R: 255, A: 255},         // Red
	{G: 255, A: 255},         // Green
	{B: 255, A: 255},         // Blue
	{R: 255, G: 255, A: 255}, // Yellow
	{R: 255, B: 255, A: 255}, // Violet
	{G: 255, B: 255, A: 255}, // Cyan
}

type TextBlock struct {
	Text       string
	Translated string
	Vertices   []*pb.Vertex
	Color      color.NRGBA
}

var errInvalidVisionPath = errors.New(`path given for Vision API service account credentials is invalid. Please run the "manga-translator-setup" application to fix it`)

// GetAnnotation gets text (TextAnnotation) from the Vision API for an image at the given file path.
func GetAnnotation(file string, url, clip bool) (*pb.TextAnnotation, error) {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Errorf("NewImageAnnotatorClient: %v", err)
		if strings.HasPrefix(err.Error(),
			"google: error getting credentials using GOOGLE_APPLICATION_CREDENTIALS environment variable") {
			return nil, errInvalidVisionPath
		}
		return nil, err
	}

	var visionImg *pb.Image
	if clip {
		imgByte := clipboard.Read(clipboard.FmtImage)
		if imgByte == nil {
			log.Fatal("Image not found in clipboard")
		}

		visionImg, err = vision.NewImageFromReader(bytes.NewReader(imgByte))
		if err != nil {
			log.Errorf("NewImageFromReader: %v", err)
			return nil, err
		}
	} else if url {
		visionImg = vision.NewImageFromURI(file)
	} else {
		f, err := os.Open(file)
		if err != nil {
			log.Errorf("Open image: %v", err)
			return nil, err
		}
		defer f.Close()
		visionImg, err = vision.NewImageFromReader(f)
		if err != nil {
			log.Errorf("NewImageFromReader: %v", err)
			return nil, err
		}
	}

	annotation, err := client.DetectDocumentText(ctx, visionImg, &pb.ImageContext{LanguageHints: []string{"ja"}})
	if err != nil {
		log.Errorf("DetectDocumentText: %v", err)
		return nil, err
	}

	if annotation == nil {
		log.Info("No text found")
		return nil, nil
	} else {
		log.WithField("text", annotation.Text).Info("Detected Text")
		return annotation, nil
	}
}

// OrganizeAnnotation converts a TextAnnotation object to a slice of TextBlocks for easier manipulation.
func OrganizeAnnotation(annotation *pb.TextAnnotation) []TextBlock {
	var blockList []TextBlock
	for _, page := range annotation.Pages {
		for i, block := range page.Blocks {
			var b string
			for _, paragraph := range block.Paragraphs {
				var p string
				for _, word := range paragraph.Words {
					symbols := make([]string, len(word.Symbols))
					for i, s := range word.Symbols {
						symbols[i] = s.Text
					}
					wordText := strings.Join(symbols, "")
					p += wordText
				}
				b += p
			}
			blockList = append(blockList, TextBlock{
				Text:     b,
				Vertices: block.BoundingBox.Vertices,
				Color:    borderColors[i%len(borderColors)], // Cycle through the list of border colors.
			})
		}
	}
	return blockList
}
