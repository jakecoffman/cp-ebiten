package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cpebiten"
	"image/color"
	"log"
	"math/rand"
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
	space.Iterations = 10
	space.SetGravity(cp.Vector{0, 100})
	space.SetCollisionSlop(0.5)

	var simpleTerrainVerts = []cp.Vector{
		{350.00, 425.07}, {336.00, 436.55}, {272.00, 435.39}, {258.00, 427.63}, {225.28, 420.00}, {202.82, 396.00},
		{191.81, 388.00}, {189.00, 381.89}, {173.00, 380.39}, {162.59, 368.00}, {150.47, 319.00}, {128.00, 311.55},
		{119.14, 286.00}, {126.84, 263.00}, {120.56, 227.00}, {141.14, 178.00}, {137.52, 162.00}, {146.51, 142.00},
		{156.23, 136.00}, {158.00, 118.27}, {170.00, 100.77}, {208.43, 84.00}, {224.00, 69.65}, {249.30, 68.00},
		{257.00, 54.77}, {363.00, 45.94}, {374.15, 54.00}, {386.00, 69.60}, {413.00, 70.73}, {456.00, 84.89},
		{468.09, 99.00}, {467.09, 123.00}, {464.92, 135.00}, {469.00, 141.03}, {497.00, 148.67}, {513.85, 180.00},
		{509.56, 223.00}, {523.51, 247.00}, {523.00, 277.00}, {497.79, 311.00}, {478.67, 348.00}, {467.90, 360.00},
		{456.76, 382.00}, {432.95, 389.00}, {417.00, 411.32}, {373.00, 433.19}, {361.00, 430.02}, {350.00, 425.07},
	}

	offset := cp.Vector{}
	for i := 0; i < len(simpleTerrainVerts)-1; i++ {
		a := simpleTerrainVerts[i]
		b := simpleTerrainVerts[i+1]
		cpebiten.AddWall(space, space.StaticBody, a.Add(offset), b.Add(offset), 0)
	}

	for i := 0; i < 1000; i++ {
		pos := randUnitCircle().Mult(180).Add(cp.Vector{screenWidth/2 + 10, screenHeight / 2})
		const radius = 5
		const mass = radius * radius / 25.0
		cpebiten.AddCircle(space, pos, mass, radius)
	}

	return &Game{
		space: space,
	}
}

func randUnitCircle() cp.Vector {
	v := cp.Vector{rand.Float64()*2.0 - 1.0, rand.Float64()*2.0 - 1.0}
	if v.LengthSq() < 1.0 {
		return v
	}
	return randUnitCircle()
}

func (g *Game) Update() error {
	cpebiten.UpdateInput(g.space)
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)

	g.space.EachShape(func(shape *cp.Shape) {
		draw := shape.UserData.(func(*ebiten.Image, *ebiten.DrawImageOptions))
		draw(screen, op)
	})

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetVsyncEnabled(false)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Benchmark")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
