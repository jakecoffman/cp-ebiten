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
	touches    map[int]*touchInfo
}

type touchInfo struct {
	id    int
	body  *cp.Body
	joint *cp.Constraint
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

	addWall(space, container, a, b)
	addWall(space, container, b, c)
	addWall(space, container, c, d)
	addWall(space, container, d, a)

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
		touches:   map[int]*touchInfo{},
	}
}

func (g *Game) handleGrab(pos cp.Vector, touchBody *cp.Body) *cp.Constraint {
	const radius = 5.0 // make it easier to grab stuff
	info := g.space.PointQueryNearest(pos, radius, Grabbable)

	// avoid infinite mass objects
	if info.Shape != nil && info.Shape.Body().Mass() < math.MaxFloat64 {
		var nearest cp.Vector
		if info.Distance > 0 {
			nearest = info.Point
		} else {
			nearest = pos
		}

		// create a joint between the invisible mouse body and the shape
		body := info.Shape.Body()
		joint := cp.NewPivotJoint2(touchBody, body, cp.Vector{}, body.WorldToLocal(nearest))
		joint.SetMaxForce(50000)
		joint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
		g.space.AddConstraint(joint)
		return joint
	}

	return nil
}

func (g *Game) Update(*ebiten.Image) error {
	x, y := ebiten.CursorPosition()
	mouse := cp.Vector{float64(x), float64(y)}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.mouseJoint = g.handleGrab(mouse, g.mouseBody)
	}
	for _, id := range ebiten.TouchIDs() {
		x, y := ebiten.TouchPosition(id)
		touchPos := cp.Vector{float64(x), float64(y)}

		touch, ok := g.touches[id]
		if !ok {
			body := cp.NewKinematicBody()
			body.SetPosition(touchPos)
			touch = &touchInfo{
				id:    id,
				body:  body,
				joint: g.handleGrab(touchPos, body),
			}
			g.touches[id] = touch
		} else {
			// lerp the touch body around which drags any shapes attached with a joint
			newPoint := touch.body.Position().Lerp(touchPos, 0.25)
			touch.body.SetVelocityVector(newPoint.Sub(touch.body.Position()).Mult(60.0))
			touch.body.SetPosition(newPoint)
		}
	}
	if g.mouseJoint != nil && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.space.RemoveConstraint(g.mouseJoint)
		g.mouseJoint = nil
	}
	for id, touch := range g.touches {
		if touch.joint != nil && inpututil.IsTouchJustReleased(id) {
			g.space.RemoveConstraint(touch.joint)
			touch.joint = nil
			delete(g.touches, id)
		}
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

	g.space.EachShape(func(shape *cp.Shape) {
		draw := shape.UserData.(func(*ebiten.Image, *ebiten.DrawImageOptions))
		draw(screen, op)
	})

	_ = ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func addWall(space *cp.Space, container *cp.Body, a, b cp.Vector) {
	seg := cp.NewSegment(container, a, b, 1).Class.(*cp.Segment)
	shape := space.AddShape(seg.Shape)
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	shape.UserData = func(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
		a, b := seg.TransformA(), seg.TransformB()
		ebitenutil.DrawLine(screen, a.X, a.Y, b.X, b.Y, color.White)
	}
}

func addBox(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
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

func addSegment(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	seg := cp.NewSegment(body,
		cp.Vector{0, (height - width) / 2.0},
		cp.Vector{0, (width - height) / 2.0},
		width/2.0).Class.(*cp.Segment)
	shape := space.AddShape(seg.Shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	a, b := seg.A(), seg.B()
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

func addCircle(space *cp.Space, pos cp.Vector, mass, radius float64) *cp.Shape {
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

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
