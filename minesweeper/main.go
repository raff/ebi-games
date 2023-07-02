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

	Count1
	Count2
	Count3
	Count4
	Count5
	Count6
	Count7
	Count8
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
	done   bool

	cw int
	ch int

	ww int
	wh int

	scale float64

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

		img, err = png.Decode(bytes.NewBuffer(pngCounts))
		if err != nil {
			log.Fatal(err)
		}

		iw, ih = img.Bounds().Dx(), img.Bounds().Dy()
		if ih != cellh || iw != cellw*8 {
			log.Fatalf("invalid cells image dimension: expected %vx%v got %vx%v", iw, ih, cellw*8, cellh)
		}

		ebimg = ebiten.NewImageFromImage(img)
		p = image.Rect(0, 0, cellw, cellh)

		for x := 0; x < iw; x += cellw {
			tile := ebimg.SubImage(p).(*ebiten.Image)
			tiles = append(tiles, tile)
			p = p.Add(image.Pt(cellw, 0))
		}
	}

	if g.ww == 0 {
		g.cw = hcount * cellw
		g.ch = vcount * cellh

		g.ww = g.cw + border + border
		g.wh = g.cw + border + border

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(background)

		g.cells = matrix.New[State](hcount, vcount, false)

		g.scale = scale

		if g.scale != 1.0 {
			g.drawOp.GeoM.Scale(g.scale, g.scale)
			g.ww = int(float64(g.ww) * g.scale)
			g.wh = int(float64(g.wh) * g.scale)
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
	g.done = false
	return g.ww, g.wh
}

func (g *Game) Coords(x, y int) (int, int) {
	x = int(float64(x) / g.scale)
	y = int(float64(y) / g.scale)

	if x < border || y < border {
		return -1, -1
	}
	if x > g.cw+border || y > g.ch+border {
		return -1, -1
	}

	return (x - border) / cellw, g.cells.Fix((y - border) / cellh)
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
			if s == Mine && !g.done {
				s = Unchecked
			}

			g.canvas.DrawImage(tiles[s], op)
			op.GeoM.Translate(cellw, 0)
		}

		op.GeoM.SetElement(0, 2, border)
		op.GeoM.Translate(0, cellh)
	}

	screen.DrawImage(g.canvas, &g.drawOp)
}

func (g *Game) countMines(x, y int) (int, []matrix.Cell[State]) {
	count := 0
	cells := g.cells.Moore(x, y, false)

	for _, c := range cells {
		if c.Value == Mine {
			count++
		}
	}

	return count, cells
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)edraw
		g.Init(0, 0, 0)

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		if g.done {
			break
		}

		x, y := g.Coords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		switch g.cells.Get(x, y) {
		case Unchecked:
			count, cells := g.countMines(x, y)

			if count > 0 {
				s := Count1 + State(count-1)
				g.cells.Set(x, y, s)
			} else {
				for _, c := range cells {
					cc, _ := g.countMines(c.X, c.Y)
					if cc > 0 {
						s := Count1 + State(cc-1)
						g.cells.Set(c.X, c.Y, s)
					} else {
						g.cells.Set(c.X, c.Y, Empty)
					}
				}

				g.cells.Set(x, y, Empty)
			}

			g.redraw = true

		case Mine:
			g.cells.Set(x, y, Exploded)
			g.redraw = true
			g.done = true
		}

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight): // Mouse click
		if g.done {
			break
		}

		x, y := g.Coords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		switch g.cells.Get(x, y) {
		case Unchecked:
			g.cells.Set(x, y, Flag)
			g.redraw = true

		case Flag:
			g.cells.Set(x, y, Unsure)
			g.redraw = true

		case Unsure:
			g.cells.Set(x, y, Unchecked)
			g.redraw = true
		}
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
