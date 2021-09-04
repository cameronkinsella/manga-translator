package window

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
)

type (
	D = layout.Dimensions
	C = layout.Context
)

// translatorWidget is the widget used for the boxes which contain the original and translated text.
func translatorWidget(gtx C, th *material.Theme, txt, title string) D {
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
			return layout.Inset{
				Top:   unit.Dp(20),
				Left:  unit.Dp(10),
				Right: unit.Dp(10)}.Layout(gtx, func(gtx C) D {

				l := material.Body1(th, txt)
				l.Font = text.Font{Typeface: "Noto"}
				l.Alignment = text.Middle
				l.Color = LightGray

				return l.Layout(gtx)
			})
		}),
	)
}

// divider is a horizontal divider widget.
func divider(gtx C) D {
	return layout.Center.Layout(gtx, func(gtx C) D {
		defer op.Save(gtx.Ops).Load()
		maxHeight := unit.Dp(4)

		d := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Px(maxHeight)}

		height := float32(gtx.Px(maxHeight))
		clip.UniformRRect(f32.Rectangle{Max: f32.Pt(float32(gtx.Constraints.Max.X), height)}, 0).Add(gtx.Ops)
		paint.ColorOp{Color: Gray}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		return D{Size: d}
	})
}

// Split is a custom layout with two widgets separated by a vertical black bar.
type Split struct{}

func (s Split) Layout(gtx C, left, right layout.Widget) D {
	bar := gtx.Px(unit.Dp(10))

	leftSize := (gtx.Constraints.Min.X + bar) / 2
	rightOffset := leftSize + bar
	rightSize := gtx.Constraints.Min.X - rightOffset

	{
		stack := op.Save(gtx.Ops)

		barRect := image.Rect(leftSize, 0, rightOffset, gtx.Constraints.Max.Y)
		clip.Rect{Max: barRect.Max, Min: barRect.Min}.Add(gtx.Ops)
		paint.ColorOp{Color: Gray}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		stack.Load()
	}

	{
		stack := op.Save(gtx.Ops)

		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftSize, gtx.Constraints.Max.Y))
		left(gtx)

		stack.Load()
	}

	{
		stack := op.Save(gtx.Ops)

		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(rightSize, gtx.Constraints.Max.Y))
		op.Offset(f32.Pt(float32(rightOffset), 0)).Add(gtx.Ops)
		right(gtx)

		stack.Load()
	}

	return D{Size: gtx.Constraints.Max}
}
