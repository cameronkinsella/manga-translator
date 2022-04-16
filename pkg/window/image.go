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
	"image"
	"math"
)

// textBlocks indicates that status of the text detection and translation process.
type textBlocks struct {
	finished bool // Is true the process is complete.
	ok       bool // Is true the process did not encounter any errors.
}

// getText performs text detection and translation for the given image and creates text block widgets for each of the text blocks.
func (t *textBlocks) getText(w *app.Window, cfg *config.File, status *string, imgInfo ImageInfo, options Options, blocks *[]detect.TextBlock, blockButtons *[]widget.Clickable) {
	// Signal goroutine death and update frame when finished.
	defer func() {
		t.finished = true
		t.ok = *status == `Done!`
		w.Invalidate()
	}()

	var blankCfg config.File

	// If the config is blank/doesn't exist, skip all steps and show error message.
	if *cfg == blankCfg {
		*status = `Your config is either blank or doesn't exist, run the "manga-translator-setup" application to create one.`
		return
	}
	// See if the block info and translations are already cached.
	*blocks = cache.Check(imgInfo.Hash, cfg.Translation.SelectedService)

	if *blocks == nil {
		*status = `Detecting text...`
		// Scan image, get text annotation.
		annotation, err := detect.GetAnnotation(imgInfo.Path, options.Url, options.Clip)
		if err != nil {
			*blocks = []detect.TextBlock{}
			*status = err.Error()
			return
		}
		*status = `Translating text...`
		// For each text block, create a new block button and add its text to allOriginal
		var allOriginal []string
		*blocks = detect.OrganizeAnnotation(annotation)
		for _, block := range *blocks {
			*blockButtons = append(*blockButtons, widget.Clickable{})
			allOriginal = append(allOriginal, block.Text)
		}

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
			*status = `Your config does not have a valid selected service, run the "manga-translator-setup" application again.`
			err = errors.New("no selected service")
		}
		for i, txt := range allTranslated {
			(*blocks)[i].Translated = txt
		}
		if err == nil {
			cache.Add(imgInfo.Hash, cfg.Translation.SelectedService, *blocks)
		} else {
			*status = allTranslated[0]
			return
		}
	} else {
		// Found in cache, we can skip annotation and translation.
		for range *blocks {
			*blockButtons = append(*blockButtons, widget.Clickable{})
		}
	}
	*status = `Done!`
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
