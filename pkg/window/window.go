package window

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/io/key"
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

var preLoadPages = 2 // hard-coded

// DrawFrame squares with labels, buttons control labels.
func DrawFrame(w *app.Window, imgInfo []ImageInfo, options Options, cfg config.File) error {

	// ops are the operations from the UI.
	var ops op.Ops

	// split is the primary application widget containing the image, translation widget, and adjustment bar.
	var split = VSplit{Ratio: 0.60}

	var pages pageList
	for _, img := range imgInfo {
		pages.append(img)
	}
	log.Debugf("Number of pages loaded: %d", pages.len)
	// Always start with first page.
	pages.current = pages.head

	// Start loading pages.
	go pages.current.load(w, &cfg, options)
	go pages.current.preLoad(preLoadPages, w, &cfg, options)

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
		selectedO string // Original text
		selectedT string // Translated text
	)

	// Listen for events in the window.
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {

			// This is sent when the application should re-render.
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				// Handle when any of the blocks are clicked.
				for i, b := range pages.current.blocks {
					if pages.current.blockButtons[i].Clicked() {
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
					return imageWidget(gtx, pages.current.text, pages.current.image, pages.current.blocks, pages.current.blockButtons)
				}, func(gtx C) D {
					return translatorPanelWidget(gtx, th, pages.current.text, originalBtn, translatedBtn, pages.current.text.status, selectedO, selectedT)
				})
				e.Frame(gtx.Ops)

			// This is sent when a key is pressed.
			case key.Event:
				if e.State == key.Press {
					if (e.Name == "→" || e.Name == "D") && pages.current.next != nil {
						pages.current = pages.current.next
						selectedO, selectedT = "", ""
						go pages.current.preLoad(preLoadPages, w, &cfg, options)
						w.Invalidate()
					} else if (e.Name == "←" || e.Name == "A") && pages.current.prev != nil {
						pages.current = pages.current.prev
						selectedO, selectedT = "", ""
						w.Invalidate()
					}
				}
			// This is sent when the application window is closed.
			case system.DestroyEvent:
				return e.Err
			}
		}
	}
}

// page is a pageList node which includes all necessary info to display an image and its translation.
type page struct {
	image        ImageInfo
	blocks       []detect.TextBlock
	blockButtons []widget.Clickable // Button widgets which will be placed over the text blocks.
	text         textBlocks
	next         *page
	prev         *page
}

// load fetches the text annotations and translations for page.
func (p *page) load(w *app.Window, cfg *config.File, options Options) {
	// Only fetch if page is not already loading or finished.
	if !p.text.loading && !p.text.finished {
		// Detect and translate text.
		p.text.getText(w, cfg, p.image, options, &p.blocks, &p.blockButtons)
	}
}

// preLoad loads the given number of pages after the current page.
func (p *page) preLoad(num int, w *app.Window, cfg *config.File, options Options) {
	cur := p.next
	for i := 0; i < num && cur != nil; i++ {
		// Asynchronously detect and translate text.
		go cur.load(w, cfg, options)
		cur = cur.next
	}
}

// pageList is a doubly linked list consisting of page objects.
type pageList struct {
	len     int
	current *page
	head    *page
	tail    *page
}

// append adds a new page with the given image info to the end of the pageList.
func (p *pageList) append(imgInfo ImageInfo) {
	newPage := &page{
		image: imgInfo,
	}
	if p.head == nil {
		p.head = newPage
		p.tail = newPage
	} else {
		currentPage := p.head
		for currentPage.next != nil {
			currentPage = currentPage.next
		}
		newPage.prev = currentPage
		currentPage.next = newPage
		p.tail = newPage
	}
	p.len++
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
