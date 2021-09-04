package image

import (
	"bytes"
	"crypto/sha256"
	log "github.com/sirupsen/logrus"
	"hash"
	"image"
	"image/draw"
	_ "image/jpeg"
	"io"
	"net/http"
	"os"
)

type Dimensions struct {
	Width, Height float32
}

// Open opens the image at the given path/URL and returns both the image and its sha256 hash.
// "url" parameter specifies if the given file string is a URL.
func Open(file string, url bool) (*image.RGBA, hash.Hash) {
	var img image.Image
	var h hash.Hash

	if url {
		resp, err := http.Get(file)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// Need TeeReader to read io.Reader twice without re-opening file
		// https://stackoverflow.com/questions/39791021/how-to-read-multiple-times-from-same-io-reader
		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)

		img, _, err = image.Decode(tee)
		if err != nil {
			log.Fatalf("Image decode error: %v", err)
		}

		h = sha256.New()
		if _, err := io.Copy(h, &buf); err != nil {
			log.Fatalf("Hash error: %v", err)
		}
	} else {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		var buf bytes.Buffer
		tee := io.TeeReader(f, &buf)

		img, _, err = image.Decode(tee)
		if err != nil {
			log.Fatalf("Image decode error: %v", err)
		}

		h = sha256.New()
		if _, err := io.Copy(h, &buf); err != nil {
			log.Fatalf("Hash error: %v", err)
		}
	}
	return convertToRGBA(img), h
}

// convertToRGBA converts the given image.Image to *image.RGBA.
func convertToRGBA(imgA image.Image) *image.RGBA {
	b := imgA.Bounds()
	imgB := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(imgB, imgB.Bounds(), imgA, b.Min, draw.Src)
	return imgB
}

// GetRatio returns the ratio multiplier for the given dimensions to conform to the given max dimension size.
func GetRatio(dims Dimensions, maxDim float32) float32 {
	var ratio float32
	if dims.Width > dims.Height {
		ratio = maxDim / dims.Width
	} else {
		ratio = maxDim / dims.Height
	}
	return ratio
}

// GetDimensions returns the dimensions of the given image.
func GetDimensions(image *image.RGBA) Dimensions {
	bounds := image.Bounds()
	return Dimensions{
		Width:  float32(bounds.Max.X),
		Height: float32(bounds.Max.Y),
	}
}
