package cpebiten

import (
	"github.com/jakecoffman/cp"
)

func AddWall(space *cp.Space, body *cp.Body, a, b cp.Vector, radius float64) *cp.Shape {
	// swap so we always draw the same direction horizontally
	if a.X < b.X {
		a, b = b, a
	}

	seg := cp.NewSegment(body, a, b, radius).Class.(*cp.Segment)
	shape := space.AddShape(seg.Shape)
	shape.SetElasticity(1)
	shape.SetFriction(1)
	shape.SetFilter(NotGrabbable)

	return shape
}

func AddSegment(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	a, b := cp.Vector{0, (height - width) / 2.0}, cp.Vector{0, (width - height) / 2.0}
	seg := cp.NewSegment(body, a, b, width/2.0).Class.(*cp.Segment)
	shape := space.AddShape(seg.Shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	return shape
}

func AddBox(space *cp.Space, pos cp.Vector, mass, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, width, height)))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewBox(body, width, height, 0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	return shape
}
func AddStaticBox(space *cp.Space, pos cp.Vector, width, height float64) *cp.Shape {
	body := space.AddBody(cp.NewKinematicBody())
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewBox(body, width, height, 0))
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	return shape
}

func AddCircle(space *cp.Space, pos cp.Vector, mass, radius float64) *cp.Shape {
	body := space.AddBody(cp.NewBody(mass, cp.MomentForCircle(mass, 0, radius, cp.Vector{})))
	body.SetPosition(pos)

	circle := cp.NewCircle(body, radius, cp.Vector{}).Class.(*cp.Circle)
	shape := space.AddShape(circle.Shape)
	shape.SetElasticity(0)
	shape.SetFriction(0.7)

	return shape
}
