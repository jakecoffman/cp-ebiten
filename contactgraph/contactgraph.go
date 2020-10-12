package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cpebiten"
	"image/color"
	"log"
)

const (
	screenWidth  = 600
	screenHeight = 480
)

type Game struct {
	space *cp.Space
	scale *cp.Body
	ball  *cp.Body
}

func NewGame() *Game {
	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{0, 300})
	space.SetCollisionSlop(0.5)
	space.SleepTimeThreshold = 1

	walls := []cp.Vector{
		{0, 0}, {0, 480},
		{0, 480}, {600, 480},
		{600, 480}, {600, 0},
	}

	for i := 0; i < len(walls)-1; i += 2 {
		cpebiten.AddWall(space, space.StaticBody, walls[i], walls[i+1], 0)
	}

	scale := cp.NewStaticBody()
	cpebiten.AddWall(space, scale, cp.Vector{50, 400}, cp.Vector{200, 400}, 4)

	for i := 0; i < 5; i++ {
		cpebiten.AddBox(space, cp.Vector{500, float64(i*32 + 220)}, 1, 30, 30)
	}

	const radius = 15
	ball := cpebiten.AddCircle(space, cp.Vector{220, 240 + radius + 5}, 10, radius).Body()

	return &Game{
		space: space,
		scale: scale,
		ball:  ball,
	}
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

	// Sum the total impulse applied to the scale from all collision pairs in the contact graph.
	var impulseSum cp.Vector
	g.scale.EachArbiter(func(arbiter *cp.Arbiter) {
		impulseSum = impulseSum.Add(arbiter.TotalImpulse())
	})

	dt := 1.0 / ebiten.CurrentTPS()

	// Force is the impulse divided by the timestep.
	force := impulseSum.Length() / dt

	// Weight can be found similarly from the gravity vector.
	gravity := g.space.Gravity()
	weight := gravity.Dot(impulseSum) / (gravity.LengthSq() * dt)

	// Highlight and count the number of shapes the ball is touching.
	var count int
	g.ball.EachArbiter(func(arb *cp.Arbiter) {
		_, other := arb.Shapes()
		cpebiten.DrawBB(screen, other.BB())
		count++
	})

	var magnitudeSum float64
	var vectorSum cp.Vector
	g.ball.EachArbiter(func(arb *cp.Arbiter) {
		j := arb.TotalImpulse()
		magnitudeSum += j.Length()
		vectorSum = vectorSum.Add(j)
	})

	crushForce := (magnitudeSum - vectorSum.Length()) * dt
	var crush string
	if crushForce > 10 {
		crush = "The ball is being crushed. (f: %.2f)"
	} else {
		crush = "The ball is not being crushed. (f %.2f)"
	}

	str := `Place objects on the scale to weigh them. The ball marks the shapes it's sitting on.
Total force: %5.2f, Total weight: %5.2f. The ball is touching %d shapes
` + crush
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf(str, force, weight, count, crushForce), 0, 100)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetVsyncEnabled(false)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Contact Graph")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
