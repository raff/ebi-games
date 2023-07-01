package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"time"

	_ "embed"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	border = 8

	cellw = 16
	cellh = 16

	hcount = 10
	vcount = 10
	nmines = 20
)

type State int

const (
	Unchecked State = iota
	Empty
	Flag
	Unsure
	UnsureChecked
	Mine
	Exploded
	Nomine
)

var (
	//go:embed assets/ms_cells.png
	pngCells []byte

	//go:embed assets/ms_counts.png
	pngCounts []byte

	tiles []*ebiten.Image

	background = color.NRGBA{80, 80, 80, 255}
)

type Game struct {
	cells matrix.Matrix[State]

	canvas *ebiten.Image
	redraw bool

	ww int
	wh int

	drawOp ebiten.DrawImageOptions
}

func (g *Game) Init(w, h int, scale float64) (int, int) {
	if len(tiles) == 0 {
		img, err := png.Decode(bytes.NewBuffer(pngCells))
		if err != nil {
			log.Fatal(err)
		}

		iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
		if ih != cellh || iw != cellw*8 {
			log.Fatalf("invalid cells image dimension: expected %vx%v got %vx%v", iw, ih, cellw*8, cellh)
		}

		ebimg := ebiten.NewImageFromImage(img)
		p := image.Rect(0, 0, cellw, cellh)

		for x := 0; x < iw; x += cellw {
			tile := ebimg.SubImage(p).(*ebiten.Image)
			tiles = append(tiles, tile)
			p = p.Add(image.Pt(cellw, 0))
		}
	}

	if g.ww == 0 {
		g.ww = hcount*cellw + border + border
		g.wh = vcount*cellh + border + border

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(background)

		g.cells = matrix.New[State](hcount, vcount, false)

		if scale != 1.0 {
			g.drawOp.GeoM.Scale(scale, scale)
			g.ww = int(float64(g.ww) * scale)
			g.wh = int(float64(g.wh) * scale)
		}
	}

	states := g.cells.Slice()
	perms := rand.Perm(len(states))

	for i, p := range perms {
		if p < nmines {
			states[i] = Mine
		} else {
			states[i] = Unchecked
		}
	}

	g.redraw = true
	return g.ww, g.wh
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(border, border)

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			s := g.cells.Get(x, y)

			g.canvas.DrawImage(tiles[s], op)
			op.GeoM.Translate(cellw, 0)
		}

		op.GeoM.SetElement(0, 2, border)
		op.GeoM.Translate(0, cellh)
	}

	screen.DrawImage(g.canvas, &g.drawOp)
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)edraw
		g.Init(0, 0, 0)
	}

	return nil
}

func main() {
	scale := flag.Float64("scale", 2, "Window scale")
	flag.Parse()

	rand.Seed(time.Now().Unix())

	g := &Game{}
	ww, wh := ebiten.ScreenSizeInFullscreen()

	ebiten.SetWindowTitle("Minesweeper")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(ww, wh, *scale))
	ebiten.RunGame(g)
}
