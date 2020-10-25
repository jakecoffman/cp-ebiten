package cpebiten

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"math"
)

const DrawPointLineScale = 1

var shader *ebiten.Shader

func init() {
	var err error
	shader, err = ebiten.NewShader([]byte(`package main
func aa_step(t1, t2, f float) float {
	return smoothstep(t1, t2, f)
}

func Fragment(position vec4, aa vec2, color vec4) vec4 {
	l := length(aa)

	fw := length(fwidth(aa))

	// Outline width threshold.
	ow := 1.0 - fw

	// Fill/outline color.
	fo_step := aa_step(max(ow - fw, 0.0), ow, l)
	fo_color := mix(color, vec4(1), fo_step)

	// Use pre-multiplied alpha.
	alpha := 1.0 - aa_step(1.0 - fw, 1.0, l)
	return fo_color*(fo_color.a*alpha)
}`))
	if err != nil {
		panic(err)
	}
}

func (o *DrawOptions) DrawBB(bb cp.BB, outline cp.FColor) {
	verts := []cp.Vector{
		{bb.R, bb.B},
		{bb.R, bb.T},
		{bb.L, bb.T},
		{bb.L, bb.B},
	}
	o.DrawPolygon(4, verts, 0, outline, cp.FColor{}, nil)
}

type DrawOptions struct {
	img *ebiten.Image

	verts   []ebiten.Vertex
	indices []uint16
	cursor  uint16
}

func NewDrawOptions(img *ebiten.Image) *DrawOptions {
	return &DrawOptions{
		img: img,
	}
}

func (o *DrawOptions) Flush() {
	o.img.DrawTrianglesShader(o.verts, o.indices, shader, &ebiten.DrawTrianglesShaderOptions{})
}

func (o *DrawOptions) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, _ interface{}) {
	r := radius + 1/DrawPointLineScale

	o.verts = append(o.verts,
		ebiten.Vertex{float32(pos.X - r), float32(pos.Y - r), -1, -1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(pos.X - r), float32(pos.Y + r), -1, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(pos.X + r), float32(pos.Y + r), 1, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(pos.X + r), float32(pos.Y - r), 1, -1, fill.R, fill.G, fill.B, fill.A},
	)
	o.indices = append(o.indices,
		o.cursor+0, o.cursor+1, o.cursor+2,
		o.cursor+0, o.cursor+2, o.cursor+3,
	)
	o.cursor += 4

	o.DrawFatSegment(pos, pos.Add(cp.ForAngle(angle).Mult(radius-DrawPointLineScale*0.5)), 0, outline, fill, nil)
}

func (o *DrawOptions) DrawSegment(a, b cp.Vector, fill cp.FColor, data interface{}) {
	o.DrawFatSegment(a, b, 0, fill, fill, data)
}

func (o *DrawOptions) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, _ interface{}) {
	n := b.Sub(a).ReversePerp().Normalize()
	t := n.ReversePerp()

	const half = 1.0 / DrawPointLineScale
	r := radius + half

	if r <= half {
		r = half
		fill = outline
	}

	nw := n.Mult(r)
	tw := t.Mult(r)
	v0 := b.Sub(nw.Add(tw))
	v1 := b.Add(nw.Sub(tw))
	v2 := b.Sub(nw)
	v3 := b.Add(nw)
	v4 := a.Sub(nw)
	v5 := a.Add(nw)
	v6 := a.Sub(nw.Sub(tw))
	v7 := a.Add(nw.Add(tw))

	o.verts = append(o.verts,
		ebiten.Vertex{float32(v0.X), float32(v0.Y), 1, -1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v1.X), float32(v1.Y), 1, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v2.X), float32(v2.Y), 0, -1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v3.X), float32(v3.Y), 0, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v4.X), float32(v4.Y), 0, -1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v5.X), float32(v5.Y), 0, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v6.X), float32(v6.Y), -1, -1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(v7.X), float32(v7.Y), -1, 1, fill.R, fill.G, fill.B, fill.A},
	)

	o.indices = append(o.indices,
		o.cursor+0, o.cursor+1, o.cursor+2,
		o.cursor+3, o.cursor+1, o.cursor+2,
		o.cursor+3, o.cursor+4, o.cursor+2,
		o.cursor+3, o.cursor+4, o.cursor+5,
		o.cursor+6, o.cursor+4, o.cursor+5,
		o.cursor+6, o.cursor+7, o.cursor+5,
	)
	o.cursor += 8
}

func (o *DrawOptions) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, _ interface{}) {
	type ExtrudeVerts struct {
		offset, n cp.Vector
	}
	extrude := make([]ExtrudeVerts, count)

	for i := 0; i < count; i++ {
		v0 := verts[(i-1+count)%count]
		v1 := verts[i]
		v2 := verts[(i+1)%count]

		n1 := v1.Sub(v0).ReversePerp().Normalize()
		n2 := v2.Sub(v1).ReversePerp().Normalize()

		offset := n1.Add(n2).Mult(1.0 / (n1.Dot(n2) + 1.0))
		extrude[i] = ExtrudeVerts{offset, n2}
	}

	inset := -math.Max(0, 1.0/DrawPointLineScale-radius)
	for i := 0; i < count-2; i++ {
		v0 := verts[0].Add(extrude[0].offset.Mult(inset))
		v1 := verts[i+1].Add(extrude[i+1].offset.Mult(inset))
		v2 := verts[i+2].Add(extrude[i+2].offset.Mult(inset))

		o.verts = append(o.verts,
			ebiten.Vertex{float32(v0.X), float32(v0.Y), 0, 0, fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(v1.X), float32(v1.Y), 0, 0, fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(v2.X), float32(v2.Y), 0, 0, fill.R, fill.G, fill.B, fill.A},
		)

		o.indices = append(o.indices, o.cursor+0, o.cursor+1, o.cursor+2)
		o.cursor += 3
	}

	outset := 1.0/DrawPointLineScale + radius - inset
	j := count - 1
	for i := 0; i < count; {
		vA := verts[i]
		vB := verts[j]

		nA := extrude[i].n
		nB := extrude[j].n

		offsetA := extrude[i].offset
		offsetB := extrude[j].offset

		innerA := vA.Add(offsetA.Mult(inset))
		innerB := vB.Add(offsetB.Mult(inset))

		inner0 := innerA
		inner1 := innerB
		outer0 := innerA.Add(nB.Mult(outset))
		outer1 := innerB.Add(nB.Mult(outset))
		outer2 := innerA.Add(offsetA.Mult(outset))
		outer3 := innerA.Add(nA.Mult(outset))

		n0 := nA
		n1 := nB
		offset0 := offsetA

		o.verts = append(o.verts,
			ebiten.Vertex{float32(inner0.X), float32(inner0.Y), 0, 0, fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(inner1.X), float32(inner1.Y), 0, 0, fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(outer0.X), float32(outer0.Y), float32(n1.X), float32(n1.Y), fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(outer1.X), float32(outer1.Y), float32(n1.X), float32(n1.Y), fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(outer2.X), float32(outer2.Y), float32(offset0.X), float32(offset0.Y), fill.R, fill.G, fill.B, fill.A},
			ebiten.Vertex{float32(outer3.X), float32(outer3.Y), float32(n0.X), float32(n0.Y), fill.R, fill.G, fill.B, fill.A},
		)

		o.indices = append(o.indices,
			o.cursor+0, o.cursor+1, o.cursor+3,
			o.cursor+0, o.cursor+2, o.cursor+3,
			o.cursor+0, o.cursor+2, o.cursor+4,
			o.cursor+0, o.cursor+4, o.cursor+5,
		)

		o.cursor += 6

		j = i
		i++
	}
}

func (o *DrawOptions) DrawDot(size float64, pos cp.Vector, fill cp.FColor, _ interface{}) {
	r := size * 0.5 / DrawPointLineScale

	o.verts = append(o.verts,
		ebiten.Vertex{float32(pos.X - r), float32(pos.Y - r), -1, -1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(pos.X - r), float32(pos.Y + r), -1, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(pos.X + r), float32(pos.Y + r), 1, 1, fill.R, fill.G, fill.B, fill.A},
		ebiten.Vertex{float32(pos.X + r), float32(pos.Y - r), 1, -1, fill.R, fill.G, fill.B, fill.A},
	)

	o.indices = append(o.indices,
		o.cursor+0, o.cursor+1, o.cursor+2,
		o.cursor+0, o.cursor+2, o.cursor+3,
	)

	o.cursor += 4
}

func (o *DrawOptions) Flags() uint {
	return 0
}

func (o *DrawOptions) OutlineColor() cp.FColor {
	return cp.FColor{
		R: 1,
		G: 0,
		B: 0,
		A: 1,
	}
}

func (o *DrawOptions) ShapeColor(shape *cp.Shape, _ interface{}) cp.FColor {
	if shape.Sensor() {
		return cp.FColor{R: 1, G: 1, B: 1, A: .1}
	}

	body := shape.Body()

	if body.IsSleeping() {
		return cp.FColor{R: .2, G: .2, B: .2, A: 1}
	}

	if body.IdleTime() > shape.Space().SleepTimeThreshold {
		return cp.FColor{R: .66, G: .66, B: .66, A: 1}
	}

	val := shape.HashId()

	// scramble the bits up using Robert Jenkins' 32 bit integer hash function
	val = (val + 0x7ed55d16) + (val << 12)
	val = (val ^ 0xc761c23c) ^ (val >> 19)
	val = (val + 0x165667b1) + (val << 5)
	val = (val + 0xd3a2646c) ^ (val << 9)
	val = (val + 0xfd7046c5) + (val << 3)
	val = (val ^ 0xb55a4f09) ^ (val >> 16)

	r := float32((val >> 0) & 0xFF)
	g := float32((val >> 8) & 0xFF)
	b := float32((val >> 16) & 0xFF)

	max := float32(math.Max(math.Max(float64(r), float64(g)), float64(b)))
	min := float32(math.Min(math.Min(float64(r), float64(g)), float64(b)))
	var intensity float32
	if body.GetType() == cp.BODY_STATIC {
		intensity = 0.15
	} else {
		intensity = 0.75
	}

	if min == max {
		return cp.FColor{R: intensity, A: 1}
	}

	var coef = intensity / (max - min)
	return cp.FColor{
		R: (r - min) * coef,
		G: (g - min) * coef,
		B: (b - min) * coef,
		A: 1,
	}
}

func (o *DrawOptions) ConstraintColor() cp.FColor {
	return cp.FColor{
		R: 0,
		G: 1,
		B: 0,
		A: 1,
	}
}

func (o *DrawOptions) CollisionPointColor() cp.FColor {
	return cp.FColor{
		R: 0,
		G: 0,
		B: 1,
		A: 1,
	}
}

func (o *DrawOptions) Data() interface{} {
	return nil
}
