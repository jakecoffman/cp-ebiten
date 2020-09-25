package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/jakecoffman/cp"
	"image/color"
	"log"
	"math"
	"math/rand"
)

const (
	screenWidth  = 600
	screenHeight = 480
)

type Game struct {
	space *cp.Space

	mouseBody  *cp.Body
	mouseJoint *cp.Constraint
}

var GrabbableMaskBit uint = 1 << 31

var Grabbable = cp.ShapeFilter{
	cp.NO_GROUP, GrabbableMaskBit, GrabbableMaskBit,
}
var NotGrabbable = cp.ShapeFilter{
	cp.NO_GROUP, ^GrabbableMaskBit, ^GrabbableMaskBit,
}

func NewGame() *Game {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{0, 600})

	container := space.AddBody(cp.NewKinematicBody())
	container.SetAngularVelocity(0.4)
	container.SetPosition(cp.Vector{300, 200})

	a := cp.Vector{-200, -200}
	b := cp.Vector{-200, 200}
	c := cp.Vector{200, 200}
	d := cp.Vector{200, -200}

	shape := space.AddShape(cp.NewSegment(container, a, b, 1))
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	shape = space.AddShape(cp.NewSegment(container, b, c, 1))
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	shape = space.AddShape(cp.NewSegment(container, c, d, 1))
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	shape = space.AddShape(cp.NewSegment(container, d, a, 1))
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	mass := 1.0
	width := 30.0
	height := width * 2

	for i := 0; i < 7; i++ {
		for j := 0; j < 3; j++ {
			pos := cp.Vector{float64(i)*width + 200, float64(j)*height + 100}

			typ := rand.Intn(3)
			if typ == 0 {
				addBox(space, pos, mass, width, height)
			} else if typ == 1 {
				addSegment(space, pos, mass, width, height)
			} else {
				addCircle(space, pos.Add(cp.Vector{0, (height - width) / 2}), mass, width/2)
				addCircle(space, pos.Add(cp.Vector{0, (width - height) / 2}), mass, width/2)
			}
		}
	}

	return &Game{
		space:     space,
		mouseBody: cp.NewKinematicBody(),
	}
}

func (g *Game) Update(*ebiten.Image) error {
	x, y := ebiten.CursorPosition()
	mouse := cp.Vector{float64(x), float64(y)}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		const radius = 5.0 // make it easier to grab stuff
		info := g.space.PointQueryNearest(mouse, radius, Grabbable)

		// avoid infinite mass objects
		if info.Shape != nil && info.Shape.Body().Mass() < math.MaxFloat64 {
			var nearest cp.Vector
			if info.Distance > 0 {
				nearest = info.Point
			} else {
				nearest = mouse
			}

			// create a joint between the invisible mouse body and the shape
			body := info.Shape.Body()
			g.mouseJoint = cp.NewPivotJoint2(g.mouseBody, body, cp.Vector{}, body.WorldToLocal(nearest))
			g.mouseJoint.SetMaxForce(50000)
			g.mouseJoint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
			g.space.AddConstraint(g.mouseJoint)
		}
	}
	if g.mouseJoint != nil && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.space.RemoveConstraint(g.mouseJoint)
		g.mouseJoint = nil
	}

	// lerp the mouse body around which drags any shapes attached with a joint
	newPoint := g.mouseBody.Position().Lerp(mouse, 0.25)
	g.mouseBody.SetVelocityVector(newPoint.Sub(g.mouseBody.Position()).Mult(60.0))
	g.mouseBody.SetPosition(newPoint)

	g.space.Step(1.0 / float64(ebiten.MaxTPS()))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	_ = screen.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)

	dc := gg.NewContext(screenWidth, screenHeight)

	g.space.EachShape(func(shape *cp.Shape) {
		op.GeoM.Reset()
		switch shape.Class.(type) {
		case *cp.Circle:
			circle := shape.Class.(*cp.Circle)

			center := circle.TransformC()
			dc.DrawCircle(center.X, center.Y, circle.Radius())
			dc.SetColor(ColorForShape(shape))
			dc.Fill()
		case *cp.Segment:
			seg := shape.Class.(*cp.Segment)

			if seg.Radius() <= 1 {
				a, b := seg.TransformA(), seg.TransformB()
				dc.DrawLine(a.X, a.Y, b.X, b.Y)
				dc.SetRGB(1, 1, 1)
				dc.Stroke()
				return
			}

			a, b := seg.A(), seg.B()
			r := seg.Radius()
			h := b.Distance(a) + r*2
			w := r * 2

			pos := seg.Body().Position()
			dc.RotateAbout(seg.Body().Angle(), pos.X, pos.Y)
			dc.DrawRoundedRectangle(pos.X-w/2, pos.Y-h/2, w, h, r)
			dc.SetColor(ColorForShape(shape))
			dc.Fill()
			dc.Identity()
		case *cp.PolyShape:
			poly := shape.Class.(*cp.PolyShape)
			n := poly.Count()

			startVert := poly.TransformVert(0)
			endVert := poly.TransformVert(n - 1)

			dc.NewSubPath()
			dc.MoveTo(startVert.X, startVert.Y)
			for i := 1; i < n-1; i++ {
				vert := poly.TransformVert(i)
				dc.LineTo(vert.X, vert.Y)
			}
			dc.LineTo(endVert.X, endVert.Y)
			dc.ClosePath()
			dc.SetColor(ColorForShape(shape))
			dc.Fill()
			dc.MoveTo(0, 0)
		}
	})

	eimg, _ := ebiten.NewImageFromImage(dc.Image(), ebiten.FilterDefault)
	_ = screen.DrawImage(eimg, op)

	_ = ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func addBox(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewBox(body, width, height, 0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
	return shape
}

func addSegment(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewSegment(body,
		cp.Vector{0, (height - width) / 2.0},
		cp.Vector{0, (width - height) / 2.0},
		width/2.0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
	return shape
}

func addCircle(space *cp.Space, pos cp.Vector, mass, radius float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
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

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
