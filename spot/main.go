package main

import (
	"bytes"
	"image"
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
	rc, rs int // card radius and symbol radius

	lhits Hits[int]
	rhits Hits[int]

	highlight *Hit[int]

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

	g.rc = max(g.tw*4, g.th*4) / 2
	g.rs = max(g.tw, g.th) / 2

	d := g.rc * 2

	g.ww = border + d + border + border + d + border
	g.wh = border + d + border

	getRect := func(x, y int) image.Rectangle {
		x = x - (g.tw / 2)
		y = y - (g.th / 2)
		return image.Rect(x, y, x+g.tw, y+g.th)
	}

	g.lhits = append(g.lhits, Hit[int]{getRect(g.rc+border, g.wh/2), 0})
	g.rhits = append(g.rhits, Hit[int]{getRect(g.ww-g.rc-border, g.wh/2), 0})

	for i := 0; i < sides; i++ {
		x := int(float64(g.rc/2) * math.Cos(2*math.Pi*float64(i)/sides))
		y := int(float64(g.rc/2) * math.Sin(2*math.Pi*float64(i)/sides))

		g.lhits = append(g.lhits, Hit[int]{getRect(g.rc+border+x, g.wh/2+y), i + 1})
		g.rhits = append(g.rhits, Hit[int]{getRect(g.ww-g.rc-border+x, g.wh/2+y), i + 1})
	}

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

	if r := g.lhits.Find(ebiten.CursorPosition()); r != nil {
		if g.highlight == nil || r.Value != g.highlight.Value {
			g.highlight = r
			g.redraw = true
		}
	} else if g.highlight != nil {
		g.highlight = nil
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
		op.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(tiles.Item(c-1), op)
	}

	vector.DrawFilledCircle(screen, float32(g.rc+border), float32(g.wh/2), float32(g.rc), color.White, false)
	vector.DrawFilledCircle(screen, float32(g.ww-g.rc-border), float32(g.wh/2), float32(g.rc), color.White, false)

	c1 := cards[0]
	c2 := cards[1]

	for i := range g.lhits {
		l := g.lhits[i]
		r := g.rhits[i]

		drawCard(l.R.Min.X, l.R.Min.Y, c1[i])
		drawCard(r.R.Min.X, r.R.Min.Y, c2[i])
	}

	if g.highlight != nil {
		r := g.highlight.R
		x := r.Dx()/2 + r.Min.X
		y := r.Dy()/2 + r.Min.Y
		vector.StrokeCircle(screen, float32(x), float32(y), float32(g.rs), 3, color.Black, false)
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
