package cpebiten

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jakecoffman/cp"
	"image/color"
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

// 8 bytes
type v2f struct {
	x, y float32
}

func V2f(v cp.Vector) v2f {
	return v2f{float32(v.X), float32(v.Y)}
}
func v2f0() v2f {
	return v2f{0, 0}
}

// 8*2 + 16*2 bytes = 48 bytes
type Vertex struct {
	vertex, aa_coord          v2f
	fill_color, outline_color cp.FColor
}

type Triangle struct {
	a, b, c Vertex
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
}

func NewDrawOptions(img *ebiten.Image) *DrawOptions {
	return &DrawOptions{
		img: img,
	}
}

func FcolorToColor(c cp.FColor) color.Color {
	return color.RGBA{
		R: uint8(c.R * 255),
		G: uint8(c.G * 255),
		B: uint8(c.B * 255),
		A: uint8(c.A * 255),
	}
}

func (o *DrawOptions) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, _ interface{}) {
	r := radius + 1/DrawPointLineScale
	a := Vertex{
		v2f{float32(pos.X - r), float32(pos.Y - r)},
		v2f{-1, -1},
		fill,
		outline,
	}
	b := Vertex{
		v2f{float32(pos.X - r), float32(pos.Y + r)},
		v2f{-1, 1},
		fill,
		outline,
	}
	c := Vertex{
		v2f{float32(pos.X + r), float32(pos.Y + r)},
		v2f{1, 1},
		fill,
		outline,
	}
	d := Vertex{
		v2f{float32(pos.X + r), float32(pos.Y - r)},
		v2f{1, -1},
		fill,
		outline,
	}

	//t0 := Triangle{a, b, c}
	//t1 := Triangle{a, c, d}

	verts := []ebiten.Vertex{
		{a.vertex.x, a.vertex.y, a.aa_coord.x, a.aa_coord.y, fill.R, fill.G, fill.B, fill.A},
		{b.vertex.x, b.vertex.y, b.aa_coord.x, b.aa_coord.y, fill.R, fill.G, fill.B, fill.A},
		{c.vertex.x, c.vertex.y, c.aa_coord.x, c.aa_coord.y, fill.R, fill.G, fill.B, fill.A},
		{d.vertex.x, d.vertex.y, d.aa_coord.x, d.aa_coord.y, fill.R, fill.G, fill.B, fill.A},
	}

	o.img.DrawTrianglesShader(verts, []uint16{0, 1, 2, 0, 2, 3}, shader, &ebiten.DrawTrianglesShaderOptions{})

	//triangleStack = append(triangleStack, t0)
	//triangleStack = append(triangleStack, t1)

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
	v0 := V2f(b.Sub(nw.Add(tw)))
	v1 := V2f(b.Add(nw.Sub(tw)))
	v2 := V2f(b.Sub(nw))
	v3 := V2f(b.Add(nw))
	v4 := V2f(a.Sub(nw))
	v5 := V2f(a.Add(nw))
	v6 := V2f(a.Sub(nw.Sub(tw)))
	v7 := V2f(a.Add(nw.Add(tw)))

	//triangles := []Triangle{{
	//	Vertex{v0, v2f{1, -1}, fill, outline},
	//	Vertex{v1, v2f{1, 1}, fill, outline},
	//	Vertex{v2, v2f{0, -1}, fill, outline},
	//},{
	//	Vertex{v3, v2f{0, 1}, fill, outline},
	//	Vertex{v1, v2f{1, 1}, fill, outline},
	//	Vertex{v2, v2f{0, -1}, fill, outline},
	//}, {
	//	Vertex{v3, v2f{0, 1}, fill, outline},
	//	Vertex{v4, v2f{0, -1}, fill, outline},
	//	Vertex{v2, v2f{0, -1}, fill, outline},
	//}, {
	//	Vertex{v3, v2f{0, 1}, fill, outline},
	//	Vertex{v4, v2f{0, -1}, fill, outline},
	//	Vertex{v5, v2f{0, 1}, fill, outline},
	//}, {
	//	Vertex{v6, v2f{-1, -1}, fill, outline},
	//	Vertex{v4, v2f{0, -1}, fill, outline},
	//	Vertex{v5, v2f{0, 1}, fill, outline},
	//}, {
	//	Vertex{v6, v2f{-1, -1}, fill, outline},
	//	Vertex{v7, v2f{-1, 1}, fill, outline},
	//	Vertex{v5, v2f{0, 1}, fill, outline},
	//}}

	verts := []ebiten.Vertex{
		{v0.x, v0.y, 1, -1, fill.R, fill.G, fill.B, fill.A},
		{v1.x, v1.y, 1, 1, fill.R, fill.G, fill.B, fill.A},
		{v2.x, v2.y, 0, -1, fill.R, fill.G, fill.B, fill.A},
		{v3.x, v3.y, 0, 1, fill.R, fill.G, fill.B, fill.A},
		{v4.x, v4.y, 0, -1, fill.R, fill.G, fill.B, fill.A},
		{v5.x, v5.y, 0, 1, fill.R, fill.G, fill.B, fill.A},
		{v6.x, v6.y, -1, -1, fill.R, fill.G, fill.B, fill.A},
		{v7.x, v7.y, -1, 1, fill.R, fill.G, fill.B, fill.A},
	}

	o.img.DrawTrianglesShader(verts, []uint16{
		0, 1, 2,
		3, 1, 2,
		3, 4, 2,
		3, 4, 5,
		6, 4, 5,
		6, 7, 5,
	}, shader, &ebiten.DrawTrianglesShaderOptions{})
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
		_ = V2f(verts[0].Add(extrude[0].offset.Mult(inset)))
		_ = V2f(verts[i+1].Add(extrude[i+1].offset.Mult(inset)))
		_ = V2f(verts[i+2].Add(extrude[i+2].offset.Mult(inset)))

		//triangleStack = append(triangleStack, Triangle{
		//	Vertex{v0, v2f0(), fill, fill},
		//	Vertex{v1, v2f0(), fill, fill},
		//	Vertex{v2, v2f0(), fill, fill},
		//})
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

		inner0 := V2f(innerA)
		inner1 := V2f(innerB)
		outer0 := V2f(innerA.Add(nB.Mult(outset)))
		outer1 := V2f(innerB.Add(nB.Mult(outset)))
		outer2 := V2f(innerA.Add(offsetA.Mult(outset)))
		outer3 := V2f(innerA.Add(nA.Mult(outset)))

		n0 := V2f(nA)
		n1 := V2f(nB)
		offset0 := V2f(offsetA)

		_ = Triangle{
			Vertex{inner0, v2f0(), fill, outline},
			Vertex{inner1, v2f0(), fill, outline},
			Vertex{outer1, n1, fill, outline},
		}
		_ = Triangle{
			Vertex{inner0, v2f0(), fill, outline},
			Vertex{outer0, n1, fill, outline},
			Vertex{outer1, n1, fill, outline},
		}
		_ = Triangle{
			Vertex{inner0, v2f0(), fill, outline},
			Vertex{outer0, n1, fill, outline},
			Vertex{outer2, offset0, fill, outline},
		}
		_ = Triangle{
			Vertex{inner0, v2f0(), fill, outline},
			Vertex{outer2, offset0, fill, outline},
			Vertex{outer3, n0, fill, outline},
		}

		j = i
		i++
	}
}

func (o *DrawOptions) DrawDot(size float64, pos cp.Vector, fill cp.FColor, _ interface{}) {
	r := size * 0.5 / DrawPointLineScale
	a := Vertex{v2f{float32(pos.X - r), float32(pos.Y - r)}, v2f{-1, -1}, fill, fill}
	b := Vertex{v2f{float32(pos.X - r), float32(pos.Y + r)}, v2f{-1, 1}, fill, fill}
	c := Vertex{v2f{float32(pos.X + r), float32(pos.Y + r)}, v2f{1, 1}, fill, fill}
	d := Vertex{v2f{float32(pos.X + r), float32(pos.Y - r)}, v2f{1, -1}, fill, fill}

	_ = Triangle{a, b, c}
	_ = Triangle{a, c, d}
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
