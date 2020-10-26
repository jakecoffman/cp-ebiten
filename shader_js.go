// +build js

package cpebiten

import "github.com/hajimehoshi/ebiten/v2"

func init() {
	var err error
	shader, err = ebiten.NewShader([]byte(`package main
func aa_step(t1, t2, f float) float {
	return smoothstep(t1, t2, f)
}

func Fragment(position vec4, aa vec2, color vec4) vec4 {
	l := length(aa)

	fw := .1

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
