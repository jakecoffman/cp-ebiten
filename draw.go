package cpebiten

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jakecoffman/cp"
	"image/color"
	"time"
)

const nanoToSec = 1_000_000_000

var currentTime float64
var lastFps = currentTime
var frames, fps int

func Draw(space *cp.Space, screen *ebiten.Image) {
	currentTime = float64(time.Now().UnixNano()) / nanoToSec
	//dt := currentTime - lastTime
	//lastTime = currentTime
	frames++
	if currentTime-lastFps >= 1 {
		fps = frames
		frames = 0
		lastFps = currentTime
	}

	screen.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)

	space.EachShape(func(shape *cp.Shape) {
		draw := shape.UserData.(func(*ebiten.Image, *ebiten.DrawImageOptions))
		draw(screen, op)
	})

	out := fmt.Sprintf("FPS: %v %0.2f", fps, ebiten.CurrentFPS())
	if profiling {
		out += "\nprofiling"
	}
	ebitenutil.DebugPrint(screen, out)
}
