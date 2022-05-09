package window

import (
	"errors"
	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	gclip "gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/cameronkinsella/manga-translator/pkg/cache"
	"github.com/cameronkinsella/manga-translator/pkg/config"
	"github.com/cameronkinsella/manga-translator/pkg/detect"
	imageW "github.com/cameronkinsella/manga-translator/pkg/image"
	"github.com/cameronkinsella/manga-translator/pkg/translate"
	log "github.com/sirupsen/logrus"
	"image"
	"math"
)

// textBlocks indicates that status of the text detection and translation process.
type textBlocks struct {
	status   string // Loading status.
	loading  bool   // Is true if the process is in progress.
	finished bool   // Is true the process is complete.
	ok       bool   // Is true the process did not encounter any errors.
}

// getText performs text detection and translation for the given image and creates text block widgets for each of the text blocks.
func (t *textBlocks) getText(w *app.Window, cfg *config.File, img imageW.TranslatorImage, blocks *[]detect.TextBlock, blockButtons *[]widget.Clickable) {
	t.loading = true

	// Signal goroutine death and update frame when finished.
	defer func() {
		t.loading = false
		t.finished = true
		t.ok = t.status == `Done!`
		w.Invalidate()
	}()

	var blankCfg config.File

	// If the config is blank/doesn't exist, skip all steps and show error message.
	if *cfg == blankCfg {
		t.status = `Your config is either blank or doesn't exist, run the "manga-translator-setup" application to create one.`
		return
	}
	var translateOnly bool
	// See if the block info and translations are already cached.
	*blocks, translateOnly = cache.Check(img.Hash, cfg.Translation.SelectedService)

	if *blocks == nil || translateOnly {
		var err error
		if !translateOnly {
			t.status = `Detecting text...`
			// Scan image, get text annotation.
			annotation, err := detect.GetAnnotation(img.Image)
			if err != nil {
				*blocks = []detect.TextBlock{}
				t.status = err.Error()
				return
			}

			*blocks = detect.OrganizeAnnotation(annotation)
		}
		// For each text block, create a new block button and add its text to allOriginal
		var allOriginal []string
		for _, block := range *blocks {
			*blockButtons = append(*blockButtons, widget.Clickable{})
			allOriginal = append(allOriginal, block.Text)
		}

		t.status = `Translating text...`
		log.Infof("Translating detected text with: %v", cfg.Translation.SelectedService)
		// Translate the text with the service specified in the config.
		var allTranslated []string
		if cfg.Translation.SelectedService == "google" {
			allTranslated, err = translate.GoogleTranslate(
				allOriginal,
				cfg.Translation.SourceLanguage,
				cfg.Translation.TargetLanguage,
				cfg.Translation.Google.APIKey,
			)
		} else if cfg.Translation.SelectedService == "deepL" {
			allTranslated, err = translate.DeepLTranslate(
				allOriginal,
				cfg.Translation.SourceLanguage,
				cfg.Translation.TargetLanguage,
				cfg.Translation.DeepL.APIKey,
			)
		} else {
			t.status = `Your config does not have a valid selected service, run the "manga-translator-setup" application again.`
			err = errors.New("no selected service")
		}
		for i, txt := range allTranslated {
			(*blocks)[i].Translated = txt
		}
		if err == nil {
			cache.Add(img.Hash, cfg.Translation.SelectedService, *blocks)
		} else {
			t.status = allTranslated[0]
			return
		}
	} else {
		// Found in cache, we can skip annotation and translation.
		for range *blocks {
			*blockButtons = append(*blockButtons, widget.Clickable{})
		}
	}
	t.status = `Done!`
}

// blockBox creates a clickable box around the given text block, and returns the widget in a StackChild.
func blockBox(img D, originalDims imageW.Dimensions, block detect.TextBlock, btn *widget.Clickable) layout.StackChild {
	return layout.Stacked(
		func(gtx C) D {
			// The vertices are for the block locations when the image is at full size.
			// To determine how much we should shrink/expand the boxes to maintain their correct position on
			// the image widget, we must find a ratio multiplier.
			ratio := imageW.GetRatio(originalDims, float32(math.Max(float64(img.Size.X), float64(img.Size.Y))))

			// Offset boxes so they start in the correct position.
			op.Offset(
				f32.Pt(float32(block.Vertices[0].X)*ratio,
					float32(block.Vertices[0].Y)*ratio),
			).Add(gtx.Ops)

			// Limit box size to ensure it stays in the area it's supposed to be in.
			boxSizeX := float32(block.Vertices[1].X - block.Vertices[0].X)
			boxSizeY := float32(block.Vertices[2].Y - block.Vertices[1].Y)
			gtx.Constraints.Max = image.Point{
				X: int(boxSizeX * ratio),
				Y: int(boxSizeY * ratio),
			}

			// Create box, filled with semi-transparent color.
			box := func(gtx C) D {
				area := gclip.Rect{
					Max: image.Point{
						X: int(boxSizeX * ratio),
						Y: int(boxSizeY * ratio),
					},
				}.Push(gtx.Ops)

				fillColor := block.Color
				fillColor.A = 0x40
				paint.ColorOp{Color: fillColor}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				defer area.Pop()
				return D{Size: gtx.Constraints.Max}
			}

			// Add opaque border around box.
			borderedBox := func(gtx C) D {
				return widget.Border{
					Color:        block.Color,
					CornerRadius: unit.Dp(1),
					Width:        unit.Dp(2),
				}.Layout(gtx, box)
			}

			return Clickable(gtx, btn, true, borderedBox)
		},
	)
}
