package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"golang.design/x/clipboard"
	drawX "golang.org/x/image/draw"
	"hash"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Dimensions struct {
	Width, Height int
}

type TranslatorImage struct {
	Image      *image.RGBA
	Hash       string
	Dimensions Dimensions
	size       int
}

// Open opens the image at the given path/URL and returns both the image and its sha256 hash.
// "url" parameter specifies if the given file string is a URL.
// "clip" parameter specifies if the image should be taken from the clipboard (overrides "url" parameter)
func Open(file string, url, clip bool) TranslatorImage {
	var img image.Image
	var h hash.Hash
	var size int

	if clip {
		var err error

		// Init returns an error if the package is not ready for use.
		err = clipboard.Init()
		if err != nil {
			log.Fatal(err)
		}

		imgByte := clipboard.Read(clipboard.FmtImage)
		if imgByte == nil {
			log.Fatal("Image not found in clipboard")
		}

		// Need TeeReader to read io.Reader twice without re-opening file
		// https://stackoverflow.com/questions/39791021/how-to-read-multiple-times-from-same-io-reader
		var buf bytes.Buffer
		tee := io.TeeReader(bytes.NewReader(imgByte), &buf)

		img, _, err = image.Decode(tee)
		size = buf.Len()
		if err != nil {
			log.Fatalf("Image decode error: %v", err)
		}

		h = sha256.New()
		if _, err := io.Copy(h, &buf); err != nil {
			log.Fatalf("Hash error: %v", err)
		}
	} else if url {
		resp, err := http.Get(file)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)

		img, _, err = image.Decode(tee)
		size = buf.Len()
		if err != nil {
			log.Fatalf("Image decode error: %v", err)
		}

		h = sha256.New()
		if _, err := io.Copy(h, &buf); err != nil {
			log.Fatalf("Hash error: %v", err)
		}
	} else {
		f, err := os.Open(filepath.ToSlash(file))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		var buf bytes.Buffer
		tee := io.TeeReader(f, &buf)

		img, _, err = image.Decode(tee)
		size = buf.Len()
		if err != nil {
			log.Fatalf("Image decode error: %v", err)
		}

		h = sha256.New()
		if _, err := io.Copy(h, &buf); err != nil {
			log.Fatalf("Hash error: %v", err)
		}
	}

	hashInBytes := h.Sum(nil)
	hashStr := hex.EncodeToString(hashInBytes)
	dims := getDimensions(img)
	imgRGBA := convertToRGBA(img)
	newImg := TranslatorImage{
		Image:      imgRGBA,
		Hash:       hashStr,
		Dimensions: dims,
		size:       size,
	}
	log.Debugf("Hash: %v", hashStr)
	log.Debugf("Image Dimensions: %v", dims)
	newImg.resize()
	return newImg
}

// convertToRGBA converts the given image.Image to *image.RGBA.
func convertToRGBA(imgA image.Image) *image.RGBA {
	b := imgA.Bounds()
	imgB := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(imgB, imgB.Bounds(), imgA, b.Min, draw.Src)
	return imgB
}

// resize reduces the given image's size to resolve file size limits and improve performance.
func (img *TranslatorImage) resize() {
	// Max size is 41943040 bytes.
	// This application is designed for images whose text is clearly visible without the need to zoom.
	// So, any images which are close to the max size will likely not lose any meaningful accuracy by reducing the image size.
	// To improve the application's performance and reduce memory usage, we will reduce the size well below the max.

	desiredSize := 20000000 // hard-coded

	if img.size <= desiredSize {
		return
	}

	log.Info("Resizing Image")
	ratio := img.size / desiredSize

	// Create new RGBA with desired dimensions.
	dst := image.NewRGBA(image.Rect(0, 0, img.Dimensions.Width/ratio, img.Dimensions.Height/ratio))

	// Scale image and draw over RGBA.
	drawX.CatmullRom.Scale(dst, dst.Rect, img.Image, img.Image.Bounds(), draw.Over, nil)
	img.Image = dst
	img.Dimensions = getDimensions(dst)
	log.Debugf("New image dimensions: %v", img.Dimensions)
}

// getSize returns the size of the given image in bytes.
func getSize(img *image.RGBA) int {
	// Create buffer.
	buff := bytes.NewBuffer(nil)

	// Encode image to buffer.
	err := png.Encode(buff, img)
	if err != nil {
		log.Fatalf("Failed to create buffer: %v", err)
	}
	return buff.Len()
}

// GetRatio returns the ratio multiplier for the given dimensions to conform to the given max dimension size.
func GetRatio(dims Dimensions, maxDim float32) float32 {
	var ratio float32
	if dims.Width > dims.Height {
		ratio = maxDim / float32(dims.Width)
	} else {
		ratio = maxDim / float32(dims.Height)
	}
	return ratio
}

// getDimensions returns the dimensions of the given image.
func getDimensions(img image.Image) Dimensions {
	bounds := img.Bounds()
	return Dimensions{
		Width:  bounds.Max.X,
		Height: bounds.Max.Y,
	}
}
