package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	border = 4
)

var (
	//go:embed assets/15puzzle.png
	tilesPng   []byte
	tiles      []*ebiten.Image
	background = color.NRGBA{64, 32, 64, 255}
)

func readTiles() (int, int) {
	img, err := png.Decode(bytes.NewBuffer(tilesPng))
	if err != nil {
		log.Fatal(err)
	}

	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	if ih != ih {
		log.Fatalf("invalid image dimension: not a square")
	}

	tw := iw / 4
	th := ih / 4

	ebimg := ebiten.NewImageFromImage(img)
	p := image.Rect(0, 0, tw, th)

	for y := 0; y < ih; y += th {
		for x := 0; x < iw; x += tw {
			tile := ebimg.SubImage(p).(*ebiten.Image)
			tiles = append(tiles, tile)
			p = p.Add(image.Pt(tw, 0))
		}

		p = p.Add(image.Pt(-iw, th))
	}

	l := len(tiles)

	// assuming the last tile is transparent (and empty) fill it with background color
	tile := tiles[l-1]
	tile.Fill(background)

	// also, move last tile (empty) to zero position
	tiles = append(tiles[l-1:], tiles[:l-1]...)

	return tw, th
}

type Game struct {
	redraw bool
	ww, wh int
	tw, th int

	cells matrix.Matrix[int]

	canvas *ebiten.Image
	drawOp ebiten.DrawImageOptions
}

func (g *Game) scramble(n int) {
	var x, y int

zero:
	for x = 0; x < g.cells.Width(); x++ {
		for y = 0; y < g.cells.Height(); y++ {
			if g.cells.Get(x, y) == 0 {
				break zero
			}
		}
	}

	for i := 0; i < n; i++ {
		// list of neighbors (up, down, left, right)
		list := g.cells.VonNewmann(x, y, false)
		cell := list[rand.Intn(len(list))] // pick random neighbor

		g.cells.Set(x, y, cell.Value)
		g.cells.Set(cell.X, cell.Y, 0)

		x, y = cell.X, cell.Y
	}
}

func (g *Game) reset() {
	l := g.cells.Slice()

	for i := range l {
		if i < len(l)-1 {
			l[i] = i + 1
		}
	}

	l[len(l)-1] = 0
}

func (g *Game) cellCoords(x, y int) (int, int) {
	if x < border || y < border {
		return -1, -1
	}
	if x > g.ww-border || y > g.wh-border {
		return -1, -1
	}

	return (x - border) / g.tw, g.cells.Fix((y - border) / g.th)
}

func (g *Game) Init(screenw, screenh int) (int, int) {
	g.tw, g.th = readTiles()
	g.ww = g.tw*4 + border + border
	g.wh = g.th*4 + border + border

	g.canvas = ebiten.NewImage(g.ww, g.wh)
	g.canvas.Fill(background)
	g.redraw = true

	g.cells = matrix.New[int](4, 4, false)
	g.reset()
	return g.ww, g.wh
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) drawCell(x, y, n int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x*g.tw+border), float64(y*g.th+border))
	g.canvas.DrawImage(tiles[n], op)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	for y := 0; y < g.cells.Height(); y++ {
		for x := 0; x < g.cells.Width(); x++ {
			v := g.cells.Get(x, y)
			g.drawCell(x, y, v)
		}
	}

	g.redraw = false
	screen.DrawImage(g.canvas, &g.drawOp)
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)andom
		g.scramble(100)
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyI): // (I)nit
		g.reset()
		g.redraw = true

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		x, y := g.cellCoords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		list := g.cells.VonNewmann(x, y, false)

		for _, cell := range list {
			if cell.Value == 0 { // swap select cell with empty one
				//log.Println("swap", x, y, "with", cell)
				g.cells.Set(cell.X, cell.Y, g.cells.Get(x, y))
				g.cells.Set(x, y, 0)
				g.redraw = true
				break
			}
		}
	}

	return nil
}

func main() {
	g := &Game{}
	ww, wh := ebiten.ScreenSizeInFullscreen()

	ebiten.SetWindowTitle("15 Puzzle")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(ww, wh))
	ebiten.RunGame(g)
}
