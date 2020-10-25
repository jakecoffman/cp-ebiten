package shader

var Circle = []byte(`
package main

var ScreenSize vec2

func circle(st vec2, r float) float {
	dist := st - vec2(0.5);
	return 1. - smoothstep(r-(r*0.01), r+(r*0.01), dot(dist,dist)*4.0);
}

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	st := position.xy/ScreenSize.xy
	col := vec3(circle(st, 0.9))
	return vec4(col, 1.0)
}
`)
