package window

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
)

// translatorWidget is the widget used for the boxes which contain the original and translated text.
func translatorWidget(gtx C, th *material.Theme, btn *widget.Clickable, txt, title string) D {
	return layout.Flex{
		Axis:      layout.Vertical,
		Spacing:   50,
		Alignment: 64}.Layout(gtx,
		layout.Rigid(divider),
		// Title
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx C) D {
				l := material.H4(th, title)
				l.Font = text.Font{Typeface: "Noto"}
				l.Alignment = text.Middle
				l.Color = LightGray

				return l.Layout(gtx)
			})
		}),
		layout.Rigid(divider),
		// Body
		layout.Rigid(func(gtx C) D {
			return Clickable(gtx, btn, false, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				w := layout.Inset{
					Top:   unit.Dp(20),
					Left:  unit.Dp(10),
					Right: unit.Dp(10)}.Layout(gtx, func(gtx C) D {

					l := material.Body1(th, txt)
					l.Font = text.Font{Typeface: "Noto"}
					l.Alignment = text.Middle
					l.Color = LightGray

					return l.Layout(gtx)
				})
				w.Size = gtx.Constraints.Max
				return w
			})
		}),
	)
}

// divider is a horizontal divider widget.
func divider(gtx C) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		maxHeight := unit.Dp(4)

		d := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Px(maxHeight)}

		height := float32(gtx.Px(maxHeight))
		area := clip.UniformRRect(f32.Rectangle{Max: f32.Pt(float32(gtx.Constraints.Max.X), height)}, 0).Push(gtx.Ops)
		paint.ColorOp{Color: Gray}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		area.Pop()

		return D{Size: d}
	})
}
