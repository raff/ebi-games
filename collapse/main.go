package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	hcount = 20
	vcount = 20
	border = 4
)

var (
	background = color.NRGBA{80, 80, 80, 255}
	//borderColor = color.NRGBA{160, 160, 160, 255}

	colors = []color.NRGBA{
		{255, 0, 0, 255},   // red
		{0, 255, 0, 255},   // green
		{0, 0, 255, 255},   // blue
		{255, 255, 0, 255}, // yellow
		{0, 255, 255, 255}, // cyan
		{255, 0, 255, 255}, // magenta
	}

	noop = &ebiten.DrawImageOptions{}

	ww, wh int // window width and height
	bw, bh int // window border
)

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func main() {
	rand.Seed(time.Now().Unix())

	g := &Game{}

	ww, wh = g.init(ebiten.ScreenSizeInFullscreen())

	ebiten.SetWindowTitle("Block Collapse")
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.RunGame(g)
}

type Game struct {
	blocks matrix.Matrix[int]
	counts [][]int

	tw, th int // game tile width, height

	canvas *ebiten.Image
}

func (g *Game) init(w, h int) (int, int) {
	if w > 0 && h > 0 {
		ww, wh = w/2, h/2

		ww = min(ww, wh)
		wh = ww

		g.tw = ww / hcount
		g.th = wh / vcount

		ww += border
		wh += border

		g.canvas = ebiten.NewImage(ww, wh)
		g.canvas.Fill(background)

		g.blocks = matrix.New[int](hcount, vcount, false)
	}

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			g.blocks.Set(x, y, rand.Intn(len(colors)))
		}
	}

	return ww, wh
}

func (g *Game) Coords(x, y int) (int, int) {
	return x / g.tw, y / g.th
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ww, wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	tile := ebiten.NewImage(g.tw-border, g.th-border)

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			color := colors[g.blocks.Get(x, y)]
			tile.Fill(color)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*g.tw+border+bw), float64(y*g.th+border+bh))
			g.canvas.DrawImage(tile, op)
		}
	}

	screen.DrawImage(g.canvas, noop)
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		g.init(0, 0)

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX):
		return fmt.Errorf("quit")

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		x, y := g.Coords(ebiten.CursorPosition())
		fmt.Println(x, y, g.blocks.Get(x, y))
	}

	return nil
}
