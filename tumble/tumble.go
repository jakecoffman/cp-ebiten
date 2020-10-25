package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cpebiten"
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
	space.SetGravity(cp.Vector{0, 600})

	container := space.AddBody(cp.NewKinematicBody())
	container.SetAngularVelocity(0.4)
	container.SetPosition(cp.Vector{screenWidth / 2, screenHeight / 2})

	a := cp.Vector{-200, -200}
	b := cp.Vector{-200, 200}
	c := cp.Vector{200, 200}
	d := cp.Vector{200, -200}

	cpebiten.AddWall(space, container, a, b, 1)
	cpebiten.AddWall(space, container, b, c, 1)
	cpebiten.AddWall(space, container, c, d, 1)
	cpebiten.AddWall(space, container, d, a, 1)

	mass := 1.0
	width := 30.0
	height := width * 2

	for i := 0; i < 7; i++ {
		for j := 0; j < 3; j++ {
			pos := cp.Vector{float64(i)*width + 200, float64(j)*height + 100}

			typ := rand.Intn(3)
			if typ == 0 {
				cpebiten.AddBox(space, pos, mass, width, height)
			} else if typ == 1 {
				cpebiten.AddSegment(space, pos, mass, width, height)
			} else {
				cpebiten.AddCircle(space, pos.Add(cp.Vector{0, (height - width) / 2}), mass, width/2)
				cpebiten.AddCircle(space, pos.Add(cp.Vector{0, (width - height) / 2}), mass, width/2)
			}
		}
	}

	return &Game{
		space: space,
	}
}

func (g *Game) Update() error {
	cpebiten.Update(g.space)
	g.space.Step(1.0 / 180.)
	g.space.Step(1.0 / 180.)
	g.space.Step(1.0 / 180.)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	opts := cpebiten.NewDrawOptions(screen)
	cp.DrawSpace(g.space, opts)
	//cpebiten.Draw(g.space, screen)
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tumble")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
