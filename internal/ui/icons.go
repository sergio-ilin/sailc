package ui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// DrawSailLogo рисует фирменный векторный логотип SailC (стилизованный геометрический парус).
func DrawSailLogo(gtx layout.Context, size unit.Dp, col color.NRGBA) layout.Dimensions {
	px := gtx.Dp(size)
	s := float32(px)

	// Координаты для двух парусов
	// Главный парус (треугольник)
	var p1 clip.Path
	p1.Begin(gtx.Ops)
	p1.MoveTo(f32.Pt(s*0.2, s*0.9))
	p1.LineTo(f32.Pt(s*0.5, s*0.1))
	p1.LineTo(f32.Pt(s*0.5, s*0.85))
	p1.Close()

	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p1.End()}.Op())

	// Малый парус (правое перо)
	var p2 clip.Path
	p2.Begin(gtx.Ops)
	p2.MoveTo(f32.Pt(s*0.58, s*0.85))
	p2.LineTo(f32.Pt(s*0.58, s*0.35))
	p2.LineTo(f32.Pt(s*0.85, s*0.85))
	p2.Close()

	// Слегка прозрачнее для глубины
	col2 := col
	col2.A = uint8(float32(col.A) * 0.75)
	paint.FillShape(gtx.Ops, col2, clip.Outline{Path: p2.End()}.Op())

	// Корпус / основание
	var p3 clip.Path
	p3.Begin(gtx.Ops)
	p3.MoveTo(f32.Pt(s*0.15, s*0.92))
	p3.LineTo(f32.Pt(s*0.85, s*0.92))
	p3.LineTo(f32.Pt(s*0.75, s*0.98))
	p3.LineTo(f32.Pt(s*0.25, s*0.98))
	p3.Close()

	paint.FillShape(gtx.Ops, col, clip.Outline{Path: p3.End()}.Op())

	return layout.Dimensions{Size: image.Pt(px, px)}
}

// DrawStatusDot рисует аккуратный круглый индикатор статуса
func DrawStatusDot(gtx layout.Context, radius unit.Dp, col color.NRGBA) layout.Dimensions {
	r := gtx.Dp(radius)
	d := r * 2
	rect := image.Rect(0, 0, d, d)
	shape := clip.UniformRRect(rect, r)
	paint.FillShape(gtx.Ops, col, shape.Op(gtx.Ops))
	return layout.Dimensions{Size: image.Pt(d, d)}
}

// DrawBadge рисует скругленный блок без рамок для тегов и статусов
func DrawBadge(gtx layout.Context, bg color.NRGBA, content layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{
		Top:    unit.Dp(3),
		Bottom: unit.Dp(3),
		Left:   unit.Dp(8),
		Right:  unit.Dp(8),
	}.Layout(gtx, content)
	call := macro.Stop()

	rect := image.Rect(0, 0, dims.Size.X, dims.Size.Y)
	paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, gtx.Dp(unit.Dp(4))).Op(gtx.Ops))
	call.Add(gtx.Ops)
	return dims
}
