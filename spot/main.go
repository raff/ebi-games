package main

import (
	"bytes"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/raff/ebi-games/util"
)

const border = 10
const sides = 7

var (
	//go:embed assets/spotit.png
	pngSyms []byte

	tiles *util.Tiles
	cards = getCards()
)

type Game struct {
	tw, th int // tile width, height
	ww, wh int // window width, height
	r      int // card radius

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

	g.tw, g.th = tiles.Width, tiles.Height

	g.r = max(g.tw*4, g.th*4) / 2

	d := g.r * 2

	g.ww = border + d + border + border + d + border
	g.wh = border + d + border
	g.redraw = true

	return g.ww, g.wh
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)edraw
		shuffle(cards)
		g.redraw = true
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	screen.Fill(color.Black)

	drawCard := func(x, y, c int) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x-(g.tw/2)), float64(y-(g.th/2)))
		screen.DrawImage(tiles.Item(c-1), op)
	}

	vector.DrawFilledCircle(screen, float32(g.r+border), float32(g.wh/2), float32(g.r), color.White, false)
	vector.DrawFilledCircle(screen, float32(g.ww-g.r-border), float32(g.wh/2), float32(g.r), color.White, false)

	c1 := cards[0]
	c2 := cards[1]

	drawCard(g.r+border, g.wh/2, c1[0])
	drawCard(g.ww-g.r-border, g.wh/2, c2[0])

	for i := 0; i < sides; i++ {
		x := int(float64(g.r/2) * math.Cos(2*math.Pi*float64(i)/sides))
		y := int(float64(g.r/2) * math.Sin(2*math.Pi*float64(i)/sides))

		drawCard(g.r+border+x, g.wh/2+y, c1[i+1])
		drawCard(g.ww-g.r-border+x, g.wh/2+y, c2[i+1])
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
