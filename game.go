package cpebiten

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"time"
)

// Game is provided as a convenience for the examples since they all share similar logic.
type Game struct {
	// Space holds all of the shapes and bodies
	Space *cp.Space

	// TicksPerSecond is the fixed physics tick rate. Set it higher if objects are going
	// through each other at the cost of higher CPU usage.
	TicksPerSecond float64

	// Accumulator shows the remaining time from the physics tick.
	Accumulator float64
	lastTime float64

	mouseBody  *cp.Body
	mouseJoint *cp.Constraint
	touches    map[ebiten.TouchID]*touchInfo

	// FixedUpdate is an optional callback that is called when a fixed update occurs.
	FixedUpdate func()
}

// NewGame creates a new game.
func NewGame(space *cp.Space, ticksPerSecond float64) *Game {
	return &Game{
		Space:          space,
		TicksPerSecond: ticksPerSecond,
		mouseBody:      cp.NewKinematicBody(),
		touches:        map[ebiten.TouchID]*touchInfo{},
	}
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
		os.Exit(0)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		if !profiling {
			f, err := os.Create("profile")
			if err != nil {
				log.Fatal(err)
			}
			profile = f
			if err := pprof.StartCPUProfile(profile); err != nil {
				log.Fatal(err)
			}
		} else {
			pprof.StopCPUProfile()
			profile.Close()
		}
		profiling = !profiling
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		ebiten.SetVsyncEnabled(vsync)
		vsync = !vsync
	}

	// web stuff
	for _, id := range inpututil.JustPressedTouchIDs() {
		x, y := ebiten.TouchPosition(id)
		touchPos := cp.Vector{float64(x), float64(y)}

		body := cp.NewKinematicBody()
		body.SetPosition(touchPos)
		touch := &touchInfo{
			id:    id,
			body:  body,
			joint: handleGrab(g.Space, touchPos, body),
		}
		g.touches[id] = touch
	}
	for id, touch := range g.touches {
		if touch.joint != nil && inpututil.IsTouchJustReleased(id) {
			g.Space.RemoveConstraint(touch.joint)
			touch.joint = nil
			delete(g.touches, id)
		} else {
			x, y := ebiten.TouchPosition(id)
			touchPos := cp.Vector{float64(x), float64(y)}
			// calculate velocity so the object goes as fast as the touch moved
			newPoint := touch.body.Position().Lerp(touchPos, 0.25)
			touch.body.SetVelocityVector(newPoint.Sub(touch.body.Position()).Mult(60.0))
			touch.body.SetPosition(newPoint)
		}
	}

	// mouse stuff
	x, y := ebiten.CursorPosition()
	if x >= 0 && y >= 0 { // fixes weird mouse stuff on mac
		mouse := cp.Vector{float64(x), float64(y)}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.mouseJoint = handleGrab(g.Space, mouse, g.mouseBody)
		}
		if g.mouseJoint != nil && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.Space.RemoveConstraint(g.mouseJoint)
			g.mouseJoint = nil
		}
		// calculate velocity so the object goes as fast as the mouse moved
		newPoint := g.mouseBody.Position().Lerp(mouse, 0.25)
		g.mouseBody.SetVelocityVector(newPoint.Sub(g.mouseBody.Position()).Mult(60.0))
		g.mouseBody.SetPosition(newPoint)
	}

	g.physicsTick()

	return nil
}

func (g *Game) physicsTick() {
	newTime := float64(time.Now().UnixNano()) / 1.e9
	frameTime := newTime - g.lastTime
	const maxUpdate = .25
	if frameTime > maxUpdate {
		frameTime = maxUpdate
	}
	g.lastTime = newTime
	g.Accumulator += frameTime

	//if !do {
	//	return
	//}

	dt := 1. / g.TicksPerSecond
	for g.Accumulator >= dt {
		if g.FixedUpdate != nil {
			g.FixedUpdate()
		}
		g.Space.Step(dt)
		g.Accumulator -= dt
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.physicsTick()

	opts := NewDrawOptions(screen)
	cp.DrawSpace(g.Space, opts)
	opts.Flush()

	out := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if profiling {
		out += "\nprofiling"
	}
	ebitenutil.DebugPrint(screen, out)
}

const (
	ScreenHeight = 480
	ScreenWidth  = 600
)

func (g *Game) Layout(int, int) (int, int) {
	return ScreenWidth, ScreenHeight
}

var GrabbableMaskBit uint = 1 << 31

var Grabbable = cp.ShapeFilter{
	cp.NO_GROUP, GrabbableMaskBit, GrabbableMaskBit,
}
var NotGrabbable = cp.ShapeFilter{
	cp.NO_GROUP, ^GrabbableMaskBit, ^GrabbableMaskBit,
}

func handleGrab(space *cp.Space, pos cp.Vector, touchBody *cp.Body) *cp.Constraint {
	const radius = 5.0 // make it easier to grab stuff
	info := space.PointQueryNearest(pos, radius, Grabbable)

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
		space.AddConstraint(joint)
		return joint
	}

	return nil
}

type touchInfo struct {
	id    ebiten.TouchID
	body  *cp.Body
	joint *cp.Constraint
}

var profiling, vsync bool
var profile *os.File
