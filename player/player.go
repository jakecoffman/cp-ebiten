package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cpebiten"
	"log"
	"math"
)

const (
	screenWidth  = 600
	screenHeight = 480
)

const (
	PlayerVelocity = 500.0

	PlayerGroundAccelTime = 0.1
	PlayerGroundAccel     = PlayerVelocity / PlayerGroundAccelTime

	PlayerAirAccelTime = 0.25
	PlayerAirAccel     = PlayerVelocity / PlayerAirAccelTime

	JumpHeight      = 50.0
	JumpBoostHeight = 55.0
	FallVelocity    = 900.0
	Gravity         = 2000.0
)

var playerBody *cp.Body
var playerShape *cp.Shape

var remainingBoost float64
var grounded, lastJumpState bool

func playerUpdateVelocity(body *cp.Body, gravity cp.Vector, damping, dt float64) {
	jumpState := ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)

	// Grab the grounding normal from last frame
	groundNormal := cp.Vector{}
	playerBody.EachArbiter(func(arb *cp.Arbiter) {
		n := arb.Normal().Neg()

		if n.Y < groundNormal.Y {
			groundNormal = n
		}
	})

	grounded = groundNormal.Y < 0
	if groundNormal.Y > 0 {
		remainingBoost = 0
	}

	// Do a normal-ish update
	boost := jumpState && remainingBoost > 0
	var g cp.Vector
	if !boost {
		g = gravity
	}
	body.UpdateVelocity(g, damping, dt)

	// Target horizontal speed for air/ground control
	var targetVx float64
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		targetVx -= PlayerVelocity
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		targetVx += PlayerVelocity
	}

	// Update the surface velocity and friction
	// Note that the "feet" move in the opposite direction of the player.
	surfaceV := cp.Vector{-targetVx, 0}
	playerShape.SetSurfaceV(surfaceV)
	if grounded {
		playerShape.SetFriction(PlayerGroundAccel / Gravity)
	} else {
		playerShape.SetFriction(0)
	}

	// Apply air control if not grounded
	if !grounded {
		v := playerBody.Velocity()
		playerBody.SetVelocity(cp.LerpConst(v.X, targetVx, PlayerAirAccel*dt), v.Y)
	}

	v := body.Velocity()
	body.SetVelocity(v.X, cp.Clamp(v.Y, -FallVelocity, cp.INFINITY))
}

type Game struct {
	space *cp.Space
}

func NewGame() *Game {
	space := cp.NewSpace()
	space.Iterations = 10
	space.SetGravity(cp.Vector{0, Gravity})

	walls := []cp.Vector{
		{0, 0}, {0, screenHeight},
		{screenWidth, 0}, {screenWidth, screenHeight},
		{0, 0}, {screenWidth, 0},
		{0, screenHeight}, {screenWidth, screenHeight},
	}
	for i := 0; i < len(walls)-1; i += 2 {
		shape := space.AddShape(cp.NewSegment(space.StaticBody, walls[i], walls[i+1], 0))
		shape.SetElasticity(1)
		shape.SetFriction(1)
		shape.SetFilter(cpebiten.NotGrabbable)
	}

	// player
	playerBody = space.AddBody(cp.NewBody(1, cp.INFINITY))
	playerBody.SetPosition(cp.Vector{100, 200})
	playerBody.SetVelocityUpdateFunc(playerUpdateVelocity)

	playerShape = space.AddShape(cp.NewBox2(playerBody, cp.BB{-15, -27.5, 15, 27.5}, 10))
	playerShape.SetElasticity(0)
	playerShape.SetFriction(0)
	playerShape.SetCollisionType(1)

	for i := 0; i < 6; i++ {
		for j := 0; j < 3; j++ {
			body := space.AddBody(cp.NewBody(4, cp.INFINITY))
			body.SetPosition(cp.Vector{float64(400 + j*60), float64(200 + i*60)})

			shape := space.AddShape(cp.NewBox(body, 50, 50, 0))
			shape.SetElasticity(0)
			shape.SetFriction(0.7)
		}
	}

	return &Game{
		space: space,
	}
}

func (g *Game) Update() error {
	jumpState := ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)

	// If the jump key was just pressed this frame, jump!
	if jumpState && !lastJumpState && grounded {
		jumpV := math.Sqrt(2.0 * JumpHeight * Gravity)
		playerBody.SetVelocityVector(playerBody.Velocity().Add(cp.Vector{0, -jumpV}))

		remainingBoost = JumpBoostHeight / jumpV
	}

	cpebiten.Update(g.space)
	g.space.Step(1.0 / 180.)
	g.space.Step(1.0 / 180.)
	g.space.Step(1.0 / 180.)

	remainingBoost -= 1./60.
	lastJumpState = jumpState

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	cpebiten.Draw(g.space, screen)
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Player")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
