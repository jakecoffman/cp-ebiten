package main

import (
	"encoding/xml"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/cpebiten"
	"golang.org/x/image/math/f64"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	_ "image/png"
)

type Game struct {
	Game    *cpebiten.Game
	map1    Map
	tileSet map[int]image.Image

	world  *ebiten.Image
	camera Camera

	drawPhysics bool
}

type Map struct {
	MapLayer [][]int // parsed tiles from the tile layer

	Width      int `xml:"width,attr"`
	Height     int `xml:"height,attr"`
	TileWidth  int `xml:"tilewidth,attr"`
	TileHeight int `xml:"tileheight,attr"`
	TileSets   []struct {
		Text     string `xml:",chardata"`
		Firstgid int    `xml:"firstgid,attr"`
		Source   string `xml:"source,attr"`
	} `xml:"tileset"`
	Layer struct {
		Text   string `xml:",chardata"`
		ID     int    `xml:"id,attr"`
		Name   string `xml:"name,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
		Data   struct {
			Text string `xml:",chardata"`
		} `xml:"data"`
	} `xml:"layer"`
	ObjectGroups struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Name   string `xml:"name,attr"`
		Object []struct {
			Text   string  `xml:",chardata"`
			ID     int     `xml:"id,attr"`
			X      float64 `xml:"x,attr"`
			Y      float64 `xml:"y,attr"`
			Width  float64 `xml:"width,attr"`
			Height float64 `xml:"height,attr"`
		} `xml:"object"`
	} `xml:"objectgroup"`
}

type TileSet struct {
	Name  string `xml:"name,attr"`
	Image struct {
		Source string `xml:"source,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
	} `xml:"image"`
}

type Camera struct {
	ViewPort   f64.Vec2
	Position   f64.Vec2
	ZoomFactor int
	Rotation   int
}

func (c *Camera) String() string {
	return fmt.Sprintf(
		"T: %.1f, R: %d, S: %d",
		c.Position, c.Rotation, c.ZoomFactor,
	)
}

func (c *Camera) viewportCenter() f64.Vec2 {
	return f64.Vec2{
		c.ViewPort[0] * 0.5,
		c.ViewPort[1] * 0.5,
	}
}

func (c *Camera) worldMatrix() ebiten.GeoM {
	m := ebiten.GeoM{}
	m.Translate(-c.Position[0], -c.Position[1])
	// We want to scale and rotate around center of image / screen
	m.Translate(-c.viewportCenter()[0], -c.viewportCenter()[1])
	m.Scale(
		math.Pow(1.01, float64(c.ZoomFactor)),
		math.Pow(1.01, float64(c.ZoomFactor)),
	)
	m.Rotate(float64(c.Rotation) * 2 * math.Pi / 360)
	m.Translate(c.viewportCenter()[0], c.viewportCenter()[1])
	return m
}

func (c *Camera) Render(world, screen *ebiten.Image) {
	screen.DrawImage(world, &ebiten.DrawImageOptions{
		GeoM: c.worldMatrix(),
	})
}

func (c *Camera) ScreenToWorld(posX, posY int) (float64, float64) {
	inverseMatrix := c.worldMatrix()
	if inverseMatrix.IsInvertible() {
		inverseMatrix.Invert()
		return inverseMatrix.Apply(float64(posX), float64(posY))
	} else {
		// When scaling it can happend that matrix is not invertable
		return math.NaN(), math.NaN()
	}
}

func (c *Camera) Reset() {
	c.Position[0] = 0
	c.Position[1] = 0
	c.Rotation = 0
	c.ZoomFactor = 0
}

func NewGame() *Game {
	f, err := os.Open("tiled/map1.tmx")
	if err != nil {
		panic(err)
	}
	var map1 Map
	if err = xml.NewDecoder(f).Decode(&map1); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}

	lines := strings.Split(strings.TrimSpace(map1.Layer.Data.Text), "\n")
	for i, line := range lines {
		map1.MapLayer = append(map1.MapLayer, []int{})
		items := strings.Split(strings.TrimSuffix(line, ","), ",")
		for _, item := range items {
			tileIndex, err := strconv.Atoi(item)
			if err != nil {
				fmt.Println("Line:", line, "Item:", item)
				panic(err)
			}
			map1.MapLayer[i] = append(map1.MapLayer[i], tileIndex)
		}
	}

	lookup := make(map[int]image.Image)

	for _, entry := range map1.TileSets {
		f, err = os.Open("tiled/" + entry.Source)
		var tileset TileSet
		if err = xml.NewDecoder(f).Decode(&tileset); err != nil {
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}

		if f, err = os.Open("tiled/" + tileset.Image.Source); err != nil {
			panic(err)
		}
		img, _, err := image.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
		tilesImage := ebiten.NewImageFromImage(img)

		index := entry.Firstgid

		for y := 0; y < tileset.Image.Height; y += map1.TileHeight {
			for x := 0; x < tileset.Image.Width; x += map1.TileWidth {
				lookup[index] = tilesImage.SubImage(image.Rect(x, y, x+map1.TileWidth, y+map1.TileHeight))
				index++
			}
		}
	}

	space := cp.NewSpace()

	for _, object := range map1.ObjectGroups.Object {
		cpebiten.AddStaticBox(space, cp.Vector{object.X + object.Width/2, object.Y + object.Height/2}, object.Width, object.Height)
	}

	worldWidth, worldHeight := map1.Width*map1.TileHeight, map1.Height*map1.TileWidth
	world := ebiten.NewImage(worldWidth, worldHeight)

	return &Game{
		Game:    cpebiten.NewGame(space, 60),
		map1:    map1,
		tileSet: lookup,
		world:   world,
		camera: Camera{
			ViewPort:   f64.Vec2{float64(worldWidth), float64(worldHeight)},
			Position:   f64.Vec2{-100, -70},
			ZoomFactor: 100,
			Rotation:   0,
		},
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.camera.Position[0] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.camera.Position[0] += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.camera.Position[1] -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.camera.Position[1] += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.camera.ZoomFactor -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.camera.ZoomFactor += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.camera.Rotation += 1
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.camera.Reset()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.drawPhysics = !g.drawPhysics
	}

	if err := g.Game.Update(); err != nil {
		return err
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.Scale(200.0/255.0, 200.0/255.0, 200.0/255.0, 1)

	for x := 0; x < g.map1.Width; x++ {
		for y := 0; y < g.map1.Height; y++ {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*g.map1.TileWidth), float64(y*g.map1.TileHeight))
			tile := g.map1.MapLayer[y][x]
			img := g.tileSet[tile]
			if img == nil {
				panic("image nil at tile " + fmt.Sprint(tile))
			}
			g.world.DrawImage(img.(*ebiten.Image), op)
		}
	}

	if g.drawPhysics {
		op := cpebiten.NewDrawOptions(g.world)
		cp.DrawSpace(g.Game.Space, op)
		op.Flush()
	}

	g.camera.Render(g.world, screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f FPS: %0.2f", ebiten.CurrentTPS(), ebiten.CurrentFPS()))
	worldX, worldY := g.camera.ScreenToWorld(ebiten.CursorPosition())
	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf("TPS: %0.2f\nMove (WASD/Arrows)\nZoom (QE)\nRotate (R)\nReset (Space)", ebiten.CurrentTPS()),
	)
	ebitenutil.DebugPrintAt(
		screen,
		fmt.Sprintf("%s\nCursor World Pos: %.2f,%.2f",
			g.camera.String(),
			worldX, worldY),
		0, screenHeight-32,
	)
}

func (g *Game) Layout(int, int) (int, int) {
	return screenWidth, screenHeight
}

const screenWidth, screenHeight = 800, 600

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tumble")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
