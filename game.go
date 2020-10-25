package cpebiten

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
	"image/color"
	"log"
	"math"
	"os"
	"runtime/pprof"
)

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

var (
	mouseBody  = cp.NewKinematicBody()
	mouseJoint *cp.Constraint
	touches    = map[ebiten.TouchID]*touchInfo{}
)

var profiling, vsync bool
var profile *os.File

func Update(space *cp.Space) {
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

	x, y := ebiten.CursorPosition()
	mouse := cp.Vector{float64(x), float64(y)}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mouseJoint = handleGrab(space, mouse, mouseBody)
	}
	for _, id := range inpututil.JustPressedTouchIDs() {
		x, y := ebiten.TouchPosition(id)
		touchPos := cp.Vector{float64(x), float64(y)}

		body := cp.NewKinematicBody()
		body.SetPosition(touchPos)
		touch := &touchInfo{
			id:    id,
			body:  body,
			joint: handleGrab(space, touchPos, body),
		}
		touches[id] = touch
	}
	for id, touch := range touches {
		if touch.joint != nil && inpututil.IsTouchJustReleased(id) {
			space.RemoveConstraint(touch.joint)
			touch.joint = nil
			delete(touches, id)
		} else {
			x, y := ebiten.TouchPosition(id)
			touchPos := cp.Vector{float64(x), float64(y)}
			// calculate velocity so the object goes as fast as the touch moved
			newPoint := touch.body.Position().Lerp(touchPos, 0.25)
			touch.body.SetVelocityVector(newPoint.Sub(touch.body.Position()).Mult(60.0))
			touch.body.SetPosition(newPoint)
		}
	}
	if mouseJoint != nil && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		space.RemoveConstraint(mouseJoint)
		mouseJoint = nil
	}

	// calculate velocity so the object goes as fast as the mouse moved
	newPoint := mouseBody.Position().Lerp(mouse, 0.25)
	mouseBody.SetVelocityVector(newPoint.Sub(mouseBody.Position()).Mult(60.0))
	mouseBody.SetPosition(newPoint)
}

func Draw(space *cp.Space, screen *ebiten.Image) {
	screen.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)

	space.EachShape(func(shape *cp.Shape) {
		draw := shape.UserData.(func(*ebiten.Image, *ebiten.DrawImageOptions))
		draw(screen, op)
	})

	out := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	if profiling {
		out += "\nprofiling"
	}
	ebitenutil.DebugPrint(screen, out)
}

