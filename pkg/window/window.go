package window

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	gclip "gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/cameronkinsella/manga-translator/pkg/config"
	"github.com/cameronkinsella/manga-translator/pkg/detect"
	imageW "github.com/cameronkinsella/manga-translator/pkg/image"
	"github.com/gonoto/notosans"
	log "github.com/sirupsen/logrus"
	"image"
	"image/color"
)

type (
	D = layout.Dimensions
	C = layout.Context
)

type ImageInfo struct {
	Image      *image.RGBA
	Path       string // Can be the local path or a URL.
	Hash       string
	Dimensions imageW.Dimensions
}

type Options struct {
	Url  bool // If the image is from a URL.
	Clip bool // If the image is in the clipboard. Takes priority over "Url" option if true.
}

// Colors
var (
	DarkGray  = color.NRGBA{R: 0x2B, G: 0x2B, B: 0x2B, A: 0xFF}
	Gray      = color.NRGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xFF}
	LightGray = color.NRGBA{R: 0xCF, G: 0xCF, B: 0xCF, A: 0xFF}
)

// DrawFrame squares with labels, buttons control labels.
func DrawFrame(w *app.Window, imgInfo ImageInfo, options Options, cfg config.File) error {

	// ops are the operations from the UI.
	var ops op.Ops

	// split is the primary application widget containing the image, translation widget, and adjustment bar.
	var split = VSplit{Ratio: 0.60}

	// Button widgets which will be placed over the text blocks.
	var blockButtons []widget.Clickable

	// Button widgets which will be placed over the translation widget for copying text to clipboard.
	var (
		originalBtn   = new(widget.Clickable)
		translatedBtn = new(widget.Clickable)
	)

	// Create material theme with Noto font to support a wide range of unicode.
	fonts := gofont.Collection()
	fonts = appendOTC(fonts, text.Font{Typeface: "Noto"}, notosans.OTC())
	th := material.NewTheme(fonts)

	var (
		txt       textBlocks
		blocks    []detect.TextBlock
		status    string // Loading status
		selectedO string // Original text
		selectedT string // Translated text
	)

	// Asynchronously detect and translate text.
	go txt.getText(w, &cfg, &status, imgInfo, options, &blocks, &blockButtons)

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
				split.Layout(gtx, func(gtx C) D {
					return imageWidget(gtx, txt, imgInfo, blocks, blockButtons)
				}, func(gtx C) D {
					return translatorPanelWidget(gtx, th, txt, originalBtn, translatedBtn, status, selectedO, selectedT)
				})
				e.Frame(gtx.Ops)

			// This is sent when the application window is closed.
			case system.DestroyEvent:
				return e.Err
			}
		}
	}
}

func imageWidget(gtx C, txt textBlocks, imgInfo ImageInfo, blocks []detect.TextBlock, blockButtons []widget.Clickable) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		imgWidget := widget.Image{
			Fit:      widget.Contain,
			Position: layout.Center,
			Src:      paint.NewImageOp(imgInfo.Image),
		}.Layout(gtx)

		// Add text blocks on top of the image widget.
		var blockWidgets []layout.StackChild

		if txt.finished {
			for i, block := range blocks {
				blockWidgets = append(blockWidgets, blockBox(imgWidget, imgInfo.Dimensions, block, &blockButtons[i]))
			}
		}

		layout.Stack{}.Layout(gtx, blockWidgets...)

		return imgWidget
	},
	)
}

// translatorPanelWidget is the full translation panel containing either the original text and translation or the current status.
func translatorPanelWidget(gtx C, th *material.Theme, txt textBlocks, originalBtn, translatedBtn *widget.Clickable, status, selectedO, selectedT string) D {
	if !txt.finished {
		return translatorWidget(gtx, th, originalBtn, status, "Loading...")
	} else if !txt.ok {
		return translatorWidget(gtx, th, originalBtn, status, "Error")
	} else {
		var tlSplit HSplit

		return tlSplit.Layout(gtx, func(gtx C) D {
			return translatorWidget(gtx, th, originalBtn, selectedO, "Original Text")
		}, func(gtx C) D {
			return translatorWidget(gtx, th, translatedBtn, selectedT, "Translated Text")
		})
	}
}

// colorBox creates a widget with the specified dimensions and color.
func colorBox(gtx C, size image.Point, color color.NRGBA) D {
	area := gclip.Rect{Max: size}.Push(gtx.Ops)
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	area.Pop()
	return D{Size: size}
}

// appendOTC adds the given OpenType font to the given font collection
func appendOTC(collection []text.FontFace, fnt text.Font, otc []byte) []text.FontFace {
	face, err := opentype.ParseCollection(otc)
	if err != nil {
		log.Fatalf("Failed to parse font collection: %v", err)
	}
	return append(collection, text.FontFace{Font: fnt, Face: face})
}
