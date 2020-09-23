package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/jakecoffman/cp"
	"image/color"
	"log"
)

const (
	screenWidth  = 600
	screenHeight = 480
)

type Game struct {
	space *cp.Space
}

func NewGame() *Game {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{0, 200})

	addBox(space, cp.Vector{100, 100}, 1, 10, 10)
	addCircle(space, cp.Vector{200, 100}, 1, 5)

	a := cp.Vector{50, 200}
	b := cp.Vector{300, 200}
	static := space.AddBody(cp.NewKinematicBody())
	space.AddShape(cp.NewSegment(static, a, b, 0))

	// We create an infinite mass rogue body to attach the line segments to
	// This way we can control the rotation however we want.
	//container := space.AddBody(cp.NewKinematicBody())
	//container.SetAngularVelocity(0.4)
	//container.SetPosition(cp.Vector{400, 200})
	//
	//a := cp.Vector{-200, 200}
	//b := cp.Vector{-200, 200}
	//c := cp.Vector{200, 200}
	//d := cp.Vector{200, -200}
	//
	//shape := space.AddShape(cp.NewSegment(container, a, b, 0))
	//shape.SetElasticity(1)
	//shape.SetFriction(1)
	////shape.SetFilter(examples.NotGrabbableFilter)
	//
	//shape = space.AddShape(cp.NewSegment(container, b, c, 0))
	//shape.SetElasticity(1)
	//shape.SetFriction(1)
	////shape.SetFilter(examples.NotGrabbableFilter)
	//
	//shape = space.AddShape(cp.NewSegment(container, c, d, 0))
	//shape.SetElasticity(1)
	//shape.SetFriction(1)
	////shape.SetFilter(examples.NotGrabbableFilter)
	//
	//shape = space.AddShape(cp.NewSegment(container, d, a, 0))
	//shape.SetElasticity(1)
	//shape.SetFriction(1)
	////shape.SetFilter(examples.NotGrabbableFilter)
	//
	//mass := 1.0
	//width := 30.0
	//height := width * 2
	//
	//for i := 0; i < 7; i++ {
	//	for j := 0; j < 3; j++ {
	//		pos := cp.Vector{float64(i)*width + 200, float64(j)*height + 100}
	//
	//		typ := rand.Intn(3)
	//		if typ == 0 {
	//			addBox(space, pos, mass, width, height)
	//		} else if typ == 1 {
	//			addSegment(space, pos, mass, width, height)
	//		} else {
	//			addCircle(space, pos.Add(cp.Vector{0, (height - width) / 2}), mass, width/2)
	//			addCircle(space, pos.Add(cp.Vector{0, (width - height) / 2}), mass, width/2)
	//		}
	//	}
	//}

	return &Game{
		space: space,
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)

	g.space.EachShape(func(shape *cp.Shape) {
		op.GeoM.Reset()
		switch shape.Class.(type) {
		case *cp.Circle:
			circle := shape.Class.(*cp.Circle)
			pos := circle.Body().Position()
			//pos := circle.TransformC()
			op.GeoM.Translate(pos.X-circle.Radius(), pos.Y-circle.Radius())
			w := int(2 * circle.Radius())
			dc := gg.NewContext(w, w)
			dc.DrawCircle(circle.Radius(), circle.Radius(), circle.Radius())
			dc.SetRGB(1, 1, 1)
			dc.Fill()
			img := dc.Image()
			eimg, _ := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
			_ = screen.DrawImage(eimg, op)
		case *cp.Segment:
			seg := shape.Class.(*cp.Segment)
			a, b := seg.A(), seg.B()
			ebitenutil.DrawLine(screen, a.X, a.Y, b.X, b.Y, color.White)
		case *cp.PolyShape:
			poly := shape.Class.(*cp.PolyShape)
			n := poly.Count()
			pos := poly.Body().Position()
			for i := 0; i < n; i++ {
				j := i + 1
				if j >= n {
					j = 0
				}
				a, b := poly.Vert(i), poly.Vert(j)
				ebitenutil.DrawLine(screen, pos.X+a.X, pos.Y+a.Y, pos.X+b.X, pos.Y+b.Y, color.White)
			}
		}
	})

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func addBox(space *cp.Space, pos cp.Vector, mass, width, height float64) {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewBox(body, width, height, 0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
}

func addSegment(space *cp.Space, pos cp.Vector, mass, width, height float64) {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewSegment(body,
		cp.Vector{0, (height - width) / 2.0},
		cp.Vector{0, (width - height) / 2.0},
		width/2.0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
}

func addCircle(space *cp.Space, pos cp.Vector, mass, radius float64) {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewCircle(body, radius, cp.Vector{}))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
