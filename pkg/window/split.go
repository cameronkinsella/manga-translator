package window

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image"
	"runtime"
)

// VSplit is a custom layout with two widgets of adjustable size separated by a draggable horizontal bar.
type VSplit struct {
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32
	// Bar is the width for resizing the layout
	Bar unit.Value

	press  bool
	hover  bool
	drag   bool
	dragID pointer.ID
	dragY  float32
}

var defaultBarWidth = unit.Dp(20) // hard-coded

func (s *VSplit) Layout(gtx C, top, bottom layout.Widget) D {
	bar := gtx.Px(s.Bar)
	if bar <= 1 {
		bar = gtx.Px(defaultBarWidth)
	}

	proportion := (s.Ratio + 1) / 2
	topSize := int(proportion*float32(gtx.Constraints.Max.Y) - float32(bar))

	bottomOffset := topSize + bar
	bottomSize := gtx.Constraints.Max.Y - bottomOffset

	// Register whole widget for release input, in case the bar is released outside its area.
	pointer.InputOp{Tag: &s,
		Types: pointer.Release,
	}.Add(gtx.Ops)

	s.handleInputs(gtx)

	{
		// Register for input.
		barRect := image.Rect(0, topSize, gtx.Constraints.Max.X, bottomOffset)
		area := clip.Rect(barRect).Push(gtx.Ops)
		paint.ColorOp{Color: LightGray}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		pointer.InputOp{Tag: s,
			Types: pointer.Press | pointer.Drag | pointer.Release | pointer.Enter | pointer.Leave,
			Grab:  s.drag,
		}.Add(gtx.Ops)
		area.Pop()

		// Add gripper lines.
		for i := 1; i <= 3; i++ {
			lineY := topSize + ((bar / 4) * i)
			gripper := image.Rect((gtx.Constraints.Max.X/2)-20, lineY+1, (gtx.Constraints.Max.X/2)+20, lineY-1)
			gripArea := clip.Rect(gripper).Push(gtx.Ops)
			paint.ColorOp{Color: DarkGray}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			gripArea.Pop()
		}
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, topSize))
		top(gtx)
	}

	{
		gtx := gtx
		off := op.Offset(f32.Pt(0, float32(bottomOffset))).Push(gtx.Ops)
		gtx.Constraints = layout.Exact(image.Pt(gtx.Constraints.Max.X, bottomSize))
		bottom(gtx)
		off.Pop()
	}

	return D{Size: gtx.Constraints.Max}
}

func (s *VSplit) handleInputs(gtx C) {
	for _, ev := range gtx.Events(s) {
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}

		switch e.Type {
		case pointer.Press:
			if s.drag {
				break
			}

			s.press = true
			s.dragID = e.PointerID
			s.dragY = e.Position.Y

		case pointer.Drag:
			if s.dragID != e.PointerID {
				break
			}

			deltaY := e.Position.Y - s.dragY
			s.dragY = e.Position.Y

			deltaRatio := deltaY * 2 / float32(gtx.Constraints.Max.Y)

			// Widget height can be max 87.5% of the window height.
			if s.Ratio+deltaRatio > 0.75 {
				s.Ratio = 0.75
			} else if s.Ratio+deltaRatio < -0.75 {
				s.Ratio = -0.75
			} else {
				s.Ratio += deltaRatio
			}
		case pointer.Enter:
			s.hover = true
		case pointer.Leave:
			s.hover = false
		case pointer.Release:
			fallthrough
		case pointer.Cancel:
			s.drag = false
			s.press = false
		}
	}

	// Define different pointers depending on OS.
	var cursorGrab pointer.Cursor
	var cursorHover pointer.Cursor
	os := runtime.GOOS
	if os == "darwin" {
		// The grab icons on macOS look nice. :)
		cursorGrab = pointer.CursorGrabbing
		cursorHover = pointer.CursorGrab
	} else {
		cursorGrab = pointer.CursorRowResize
		cursorHover = pointer.CursorRowResize
	}

	if s.press {
		cursorGrab.Add(gtx.Ops)
	} else if s.hover {
		cursorHover.Add(gtx.Ops)
	} else {
		pointer.CursorDefault.Add(gtx.Ops)
	}
}

// HSplit is a custom layout with two widgets of equal size separated by a vertical bar.
type HSplit struct{}

func (s *HSplit) Layout(gtx C, left, right layout.Widget) D {
	bar := gtx.Px(unit.Dp(10))

	leftSize := (gtx.Constraints.Min.X - bar) / 2
	rightOffset := leftSize + bar
	rightSize := gtx.Constraints.Min.X - rightOffset

	{
		barRect := image.Rect(leftSize, 0, rightOffset, gtx.Constraints.Max.Y)
		area := clip.Rect{Max: barRect.Max, Min: barRect.Min}.Push(gtx.Ops)
		paint.ColorOp{Color: Gray}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		area.Pop()
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftSize, gtx.Constraints.Max.Y))
		left(gtx)
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(rightSize, gtx.Constraints.Max.Y))
		op.Offset(f32.Pt(float32(rightOffset), 0)).Add(gtx.Ops)
		right(gtx)
	}

	return D{Size: gtx.Constraints.Max}
}
