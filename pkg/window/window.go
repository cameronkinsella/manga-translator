package window

import (
	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cameronkinsella/manga-translator/pkg/cache"
	"github.com/cameronkinsella/manga-translator/pkg/config"
	"github.com/cameronkinsella/manga-translator/pkg/detect"
	imageW "github.com/cameronkinsella/manga-translator/pkg/image"
	"github.com/cameronkinsella/manga-translator/pkg/translate"
	"github.com/gonoto/notosans"
	log "github.com/sirupsen/logrus"
	"image"
	"image/color"
	"math"
)

// Colors
var (
	DarkGray  = color.NRGBA{R: 0x2B, G: 0x2B, B: 0x2B, A: 0xFF}
	Gray      = color.NRGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xFF}
	LightGray = color.NRGBA{R: 0xCF, G: 0xCF, B: 0xCF, A: 0xFF}
)

// DrawFrame squares with labels, buttons control labels. "url" states if the given imgPath is a URL or not.
func DrawFrame(w *app.Window, img *image.RGBA, imgPath string, imgHash string, imgDims imageW.Dimensions, url bool, cfg config.File) error {

	// ops are the operations from the UI
	var ops op.Ops

	// Button widgets which will be placed over the text blocks.
	var blockButtons []widget.Clickable

	// Button widgets which will be placed over the translation widget for copying text to clipboard.
	var originalBtn, translatedBtn widget.Clickable

	// Create material theme with Noto font to support a wide range of unicode.
	fonts := gofont.Collection()
	fonts = appendOTC(fonts, text.Font{Typeface: "Noto"}, notosans.OTC())
	th := material.NewTheme(fonts)

	var (
		blocks    []detect.TextBlock
		selectedO string // Original text
		selectedT string // Translated text
	)

	var blankCfg config.File

	// If the config is blank/doesn't exist, skip all steps and
	if cfg == blankCfg {
		selectedO = `Your config is either blank or doesn't exist, run the "config-setup" application to create one.`
	} else {
		// See if the block info and translations are already cached.
		blocks = cache.Check(imgHash, cfg.Translation.SelectedService)

		if blocks == nil {
			// Scan image, get text annotation.
			annotation, err := detect.GetAnnotation(imgPath, url)
			if err != nil {
				blocks = []detect.TextBlock{}
				selectedO = err.Error()
			} else {

				// For each text block, create a new block button and add its text to allOriginal
				var allOriginal []string
				blocks = detect.OrganizeAnnotation(annotation)
				for _, block := range blocks {
					blockButtons = append(blockButtons, widget.Clickable{})
					allOriginal = append(allOriginal, block.Text)
				}

				// Translate the text with the service specified in the config.
				var allTranslated []string
				if cfg.Translation.SelectedService == "google" {
					allTranslated = translate.GoogleTranslate(allOriginal, cfg)
				} else if cfg.Translation.SelectedService == "deepL" {
					allTranslated = translate.DeepLTranslate(allOriginal, cfg)
				} else {
					allTranslated = translate.TranslationError(
						`Your config does not have a valid selected service, run the "config-setup" application again.`,
						allOriginal,
					)
				}
				for i, txt := range allTranslated {
					blocks[i].Translated = txt
				}
				if cfg.Translation.SelectedService == "google" || cfg.Translation.SelectedService == "deepL" {
					cache.Add(imgHash, cfg.Translation.SelectedService, blocks)
				}
			}
		} else {
			for range blocks {
				blockButtons = append(blockButtons, widget.Clickable{})
			}
		}
	}

	// Listen for events in the window.
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {

			// This is sent when the application should re-render.
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				// Handle when any of the blocks are clicked.
				for i, b := range blocks {
					if blockButtons[i].Clicked() {
						log.Debugf("Clicked Block %d", i)
						selectedO = b.Text
						selectedT = b.Translated
					}
				}

				// Write to clipboard if either of the text sections are clicked.
				if originalBtn.Clicked() {
					w.WriteClipboard(selectedO)
				} else if translatedBtn.Clicked() {
					w.WriteClipboard(selectedT)
				}

				// Background
				layout.Center.Layout(gtx, func(gtx C) D {
					return colorBox(gtx, gtx.Constraints.Max, DarkGray)
				})

				// Application
				layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceEnd,
				}.Layout(gtx,
					// Image
					layout.Rigid(func(gtx C) D {
						return layout.Center.Layout(gtx, func(gtx C) D {
							gtx.Constraints.Max.Y -= 200
							imgWidget := widget.Image{
								Fit:      widget.Contain,
								Position: layout.Center,
								Src:      paint.NewImageOp(img),
							}.Layout(gtx)

							// Add text blocks on top of the image widget.
							var blockWidgets []layout.StackChild

							for i, block := range blocks {
								blockWidgets = append(blockWidgets, blockBox(imgWidget, imgDims, block, &blockButtons[i]))
							}

							layout.Stack{}.Layout(gtx, blockWidgets...)

							return imgWidget
						},
						)
					}),
					// Translation panel
					layout.Rigid(
						func(gtx C) D {
							var split Split

							return split.Layout(gtx, func(gtx C) D {
								return translatorWidget(gtx, th, &originalBtn, selectedO, "Original Text")
							}, func(gtx C) D {
								return translatorWidget(gtx, th, &translatedBtn, selectedT, "Translated Text")
							})
						},
					),
				)
				e.Frame(gtx.Ops)

			// This is sent when the application window is closed.
			case system.DestroyEvent:
				return e.Err
			}
		}
	}
}

// blockBox creates a clickable box around the given text block, and returns the widget in a StackChild.
func blockBox(img D, originalDims imageW.Dimensions, block detect.TextBlock, btn *widget.Clickable) layout.StackChild {
	return layout.Stacked(
		func(gtx C) D {
			defer op.Save(gtx.Ops).Load()
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
				defer op.Save(gtx.Ops).Load()
				clip.Rect{
					Max: image.Point{
						X: int(boxSizeX * ratio),
						Y: int(boxSizeY * ratio),
					},
				}.Add(gtx.Ops)

				fillColor := block.Color
				fillColor.A = 0x40
				paint.ColorOp{Color: fillColor}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
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

			return material.Clickable(gtx, btn, borderedBox)
		},
	)
}

// colorBox creates a widget with the specified dimensions and color.
func colorBox(gtx layout.Context, size image.Point, color color.NRGBA) layout.Dimensions {
	defer op.Save(gtx.Ops).Load()
	clip.Rect{Max: size}.Add(gtx.Ops)
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

// appendOTC adds the given OpenType font to the given font collection
func appendOTC(collection []text.FontFace, fnt text.Font, otc []byte) []text.FontFace {
	face, err := opentype.ParseCollection(otc)
	if err != nil {
		log.Fatalf("Failed to parse font collection: %v", err)
	}
	return append(collection, text.FontFace{Font: fnt, Face: face})
}
