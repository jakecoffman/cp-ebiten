package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cpebiten"
	"image/color"
	"log"
)

const (
	screenWidth  = 600
	screenHeight = 480
)

const (
	chainCount = 8
	linkCount  = 10
)

func BreakableJointPostStepRemove(space *cp.Space, joint interface{}, _ interface{}) {
	space.RemoveConstraint(joint.(*cp.Constraint))
}

func BreakableJointPostSolve(joint *cp.Constraint, space *cp.Space) {
	dt := space.TimeStep()

	// Convert the impulse to a force by dividing it by the timestep.
	force := joint.Class.GetImpulse() / dt
	maxForce := joint.MaxForce()

	// If the force is almost as big as the joint's max force, break it.
	if force > 0.9*maxForce {
		space.AddPostStepCallback(BreakableJointPostStepRemove, joint, nil)
	}
}

type Game struct {
	space *cp.Space
}

func NewGame() *Game {
	space := cp.NewSpace()
	space.Iterations = 30
	space.SetGravity(cp.Vector{0, 100})
	space.SleepTimeThreshold = 0.5

	walls := []cp.Vector{
		{0, 0}, {0, screenHeight},
		{0, screenHeight}, {screenWidth, screenHeight},
		{screenWidth, screenHeight}, {screenWidth, 0},
		{screenWidth, 0}, {0, 0},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		cpebiten.AddWall(space, space.StaticBody, walls[i], walls[i+1], 0)
	}

	mass := 1.0
	width := 20.0
	height := 30.0

	spacing := width * 0.3

	var i, j float64
	for i = 0; i < chainCount; i++ {
		var prev *cp.Body

		for j = 0; j < linkCount; j++ {
			pos := cp.Vector{
				X: screenWidth/2 + 40 * (i - (chainCount-1)/2.0),
				Y: (j+0.5)*height + (j+1)*spacing,
			}

			shape := cpebiten.AddSegment(space, pos, mass, width, height)

			breakingForce := 80000.0

			var constraint *cp.Constraint
			if prev == nil {
				a, b := cp.Vector{0, -height / 2}, cp.Vector{pos.X, 0}
				constraint = space.AddConstraint(cp.NewSlideJoint(shape.Body(), space.StaticBody, a, b, 0, spacing))
			} else {
				a, b := cp.Vector{0, - height / 2}, cp.Vector{0, height / 2}
				constraint = space.AddConstraint(cp.NewSlideJoint(shape.Body(), prev, a, b, 0, spacing))
			}

			constraint.SetMaxForce(breakingForce)
			constraint.PostSolve = BreakableJointPostSolve
			constraint.SetCollideBodies(false)

			prev = shape.Body()
		}
	}

	radius := 15.0
	circle := cpebiten.AddCircle(space, cp.Vector{screenWidth/2, screenHeight - 100}, 10, radius)
	circle.Body().SetVelocity(0, -300)

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
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Contact Graph")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
