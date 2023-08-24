package main

import (
	"bytes"
	"image/color"
	"log"
	"math/rand"
	"time"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	//"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/raff/ebi-games/util"
)

const border = 4

var (
	//go:embed assets/spotit.png
	pngSyms []byte

	tiles *util.Tiles
)

type Game struct {
	ww, wh int

	redraw bool
}

func (g *Game) Init() (int, int) {
	if tiles == nil {
		var err error

		tiles, err = util.ReadTiles(bytes.NewBuffer(pngSyms), 10, 6)
		if err != nil {
			log.Fatal(err)
		}
	}

	tw, th := tiles.Width+border, tiles.Height+border

	g.ww = tw*tiles.Columns + border
	g.wh = th*tiles.Rows + border
	g.redraw = true

	return g.ww, g.wh
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	screen.Fill(color.Black)

	tw := tiles.Width + border
	th := tiles.Height + border

	filler := ebiten.NewImage(tiles.Width, tiles.Height)
	filler.Fill(color.White)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(border), float64(border))

	for y := 0; y < tiles.Rows; y++ {
		for x := 0; x < tiles.Columns; x++ {
			t := tiles.At(x, y)
			screen.DrawImage(filler, op)
			screen.DrawImage(t, op)
			op.GeoM.Translate(float64(tw), 0)
		}

		op.GeoM.SetElement(0, 2, border)
		op.GeoM.Translate(0, float64(th))
	}

	g.redraw = false
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ww, g.wh
}

func main() {
	rand.Seed(time.Now().Unix())

	g := &Game{}

	ebiten.SetWindowTitle("Spot!")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init())
	ebiten.RunGame(g)
}
