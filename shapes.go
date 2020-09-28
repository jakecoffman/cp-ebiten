package cpebiten

import (
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/jakecoffman/cp"
	"image/color"
	"math"
)

func AddWall(space *cp.Space, body *cp.Body, a, b cp.Vector, radius float64) *cp.Shape {
	// swap so we always draw the same direction horizontally
	if a.X < b.X {
		a, b = b, a
	}

	seg := cp.NewSegment(body, a, b, radius).Class.(*cp.Segment)
	shape := space.AddShape(seg.Shape)
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	r := seg.Radius()
	if r == 0 {
		r = 1
	}
	h := b.Distance(a) + r*2
	w := r * 2

	dc := gg.NewContext(int(w), int(h))
	dc.DrawRoundedRectangle(0, 0, w, h, r)
	dc.SetRGB(1, 1, 1)
	dc.Fill()
	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)

	center := cp.Vector{(a.X + b.X)/2, (a.Y+b.Y)/2}
	offset := center.Sub(body.Position())

	maxY := math.Max(a.Y, b.Y)
	minY := math.Min(a.Y, b.Y)
	rotation := math.Acos((maxY - minY) / math.Sqrt(math.Pow(a.X-b.X, 2)+math.Pow(maxY-minY, 2)))
	if b.Y < a.Y {
		rotation *= -1
	}

	shape.UserData = func(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
		op.GeoM.Translate(-w/2, -h/2)
		op.GeoM.Rotate(rotation)
		op.GeoM.Translate(offset.X, offset.Y)

		pos := seg.Body().Position()
		op.GeoM.Translate(pos.X, pos.Y)
		op.GeoM.Rotate(seg.Body().Angle())
		op.GeoM.Translate(pos.X, pos.Y)

		_ = screen.DrawImage(img, op)
		op.GeoM.Reset()
	}
	return shape
}

func AddSegment(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	a, b := cp.Vector{0, (height - width) / 2.0}, cp.Vector{0, (width - height) / 2.0}
	seg := cp.NewSegment(body, a, b, width/2.0).Class.(*cp.Segment)
	shape := space.AddShape(seg.Shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	r := seg.Radius()
	h := b.Distance(a) + r*2
	w := r * 2

	dc := gg.NewContext(int(width), int(height))
	dc.DrawRoundedRectangle(0, 0, w, h, r)
	dc.SetColor(ColorForShape(shape))
	dc.Fill()
	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)

	shape.UserData = func(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
		op.GeoM.Translate(-width/2, -height/2)
		op.GeoM.Rotate(seg.Body().Angle())
		pos := seg.Body().Position()
		op.GeoM.Translate(pos.X, pos.Y)
		_ = screen.DrawImage(img, op)
		op.GeoM.Reset()
	}
	return shape
}

func AddBox(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewBox(body, width, height, 0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	dc := gg.NewContext(int(width), int(height))
	dc.DrawRectangle(0, 0, width, height)
	dc.SetColor(ColorForShape(shape))
	dc.Fill()
	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)

	shape.UserData = func(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
		op.GeoM.Translate(-width/2, -height/2)
		op.GeoM.Rotate(body.Angle())
		pos := body.Position()
		op.GeoM.Translate(pos.X, pos.Y)
		_ = screen.DrawImage(img, op)
		op.GeoM.Reset()
	}

	return shape
}

func AddCircle(space *cp.Space, pos cp.Vector, mass, radius float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	circle := cp.NewCircle(body, radius, cp.Vector{}).Class.(*cp.Circle)
	shape := space.AddShape(circle.Shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	dc := gg.NewContext(int(circle.Radius()*2), int(circle.Radius()*2))
	dc.DrawCircle(circle.Radius(), circle.Radius(), circle.Radius())
	dc.SetColor(ColorForShape(shape))
	dc.Fill()
	img, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)
	shape.UserData = func(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
		op.GeoM.Translate(-circle.Radius(), -circle.Radius())
		op.GeoM.Rotate(body.Angle())
		center := circle.TransformC()
		op.GeoM.Translate(center.X, center.Y)
		_ = screen.DrawImage(img, op)
		op.GeoM.Reset()
	}

	return shape
}

func ColorForShape(shape *cp.Shape) color.Color {
	if shape.Sensor() {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255 / 10}
	}

	body := shape.Body()

	if body.IsSleeping() {
		return color.NRGBA{R: 255 / 20, G: 255 / 20, B: 255 / 20, A: 255}
	}

	if body.IdleTime() > shape.Space().SleepTimeThreshold {
		return color.NRGBA{R: 255 * 2 / 3, G: 255 * 2 / 3, B: 255 * 2 / 3, A: 255}
	}

	val := shape.HashId()

	// scramble the bits up using Robert Jenkins' 32 bit integer hash function
	val = (val + 0x7ed55d16) + (val << 12)
	val = (val ^ 0xc761c23c) ^ (val >> 19)
	val = (val + 0x165667b1) + (val << 5)
	val = (val + 0xd3a2646c) ^ (val << 9)
	val = (val + 0xfd7046c5) + (val << 3)
	val = (val ^ 0xb55a4f09) ^ (val >> 16)

	r := float32((val >> 0) & 0xFF)
	g := float32((val >> 8) & 0xFF)
	b := float32((val >> 16) & 0xFF)

	max := float32(math.Max(math.Max(float64(r), float64(g)), float64(b)))
	min := float32(math.Min(math.Min(float64(r), float64(g)), float64(b)))
	var intensity float32
	if body.GetType() == cp.BODY_STATIC {
		intensity = 0.15
	} else {
		intensity = 0.75
	}

	if min == max {
		return color.NRGBA{R: uint8(255 * intensity), A: 1}
	}

	coef := intensity / (max - min)
	return color.NRGBA{
		R: uint8(255 * (r - min) * coef),
		G: uint8(255 * (g - min) * coef),
		B: uint8(255 * (b - min) * coef),
		A: 255,
	}
}

func DrawBB(image *ebiten.Image, bb cp.BB) {
	red := color.RGBA{255, 0, 0, 255}
	ebitenutil.DrawLine(image, bb.R, bb.B, bb.R, bb.T, red)
	ebitenutil.DrawLine(image, bb.R, bb.T, bb.L, bb.T, red)
	ebitenutil.DrawLine(image, bb.L, bb.T, bb.L, bb.B, red)
	ebitenutil.DrawLine(image, bb.L, bb.B, bb.R, bb.B, red)
}
