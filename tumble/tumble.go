package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
	space.SetGravity(cp.Vector{0, 600})

	container := space.AddBody(cp.NewKinematicBody())
	container.SetAngularVelocity(0.4)
	container.SetPosition(cp.Vector{screenWidth/2, screenHeight/2})

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

func (g *Game) Update(*ebiten.Image) error {
	cpebiten.UpdateInput(g.space)
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

	_ = ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetVsyncEnabled(false)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebiten")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
