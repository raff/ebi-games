package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"sort"
	"time"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	hcount = 25
	vcount = 25
	border = 4

	nmatch  = 3
	nrefill = 5

	visited = -1
	empty   = -2
	bg      = -3
	high    = -4

	title = "Block Collapse"
)

var (
	background     = color.NRGBA{80, 80, 80, 255}
	highlightColor = color.NRGBA{240, 220, 240, 255}

	colors = []color.NRGBA{
		{255, 0, 0, 255},     // red
		{0, 255, 0, 255},     // green
		{0, 0, 255, 255},     // blue
		{255, 255, 0, 255},   // yellow
		{128, 255, 255, 255}, // cyan
		{255, 128, 255, 255}, // magenta
	}

	noop = &ebiten.DrawImageOptions{}

	gomessage = []int{
		1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 1, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1, 0,
		0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0,
		0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0, 0,
		0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0,
		0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0,
		0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 0, 0,
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0,
		0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0,
		0, 1, 0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0,
		0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0,
		0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 0,
		0, 0, 1, 1, 1, 0, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0,
		1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	gow = 24
	goh = 15
)

func main() {
	rand.Seed(time.Now().Unix())

	g := &Game{}

	ebiten.SetWindowTitle(title)
	//ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.SetWindowSize(g.Init(ebiten.ScreenSizeInFullscreen()))
	ebiten.RunGame(g)
}

type Point struct {
	x, y int
}

type Game struct {
	blocks matrix.Matrix[int]

	ww, wh int // window width, height
	tw, th int // game tile width, height

	highlight []Point
	autoplay  bool

	score int

	canvas *ebiten.Image // image buffer
}

func (g *Game) Init(w, h int) (int, int) {
	if w > 0 && h > 0 {
		g.ww, g.wh = w/2, h/2
		if g.ww < g.wh {
			g.wh = g.ww
		} else {
			g.ww = g.wh
		}

		g.tw = g.ww / hcount
		g.th = g.wh / vcount

		g.ww += border
		g.wh += border

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(background)

		g.blocks = matrix.New[int](hcount, vcount, false)
	}

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			g.blocks.Set(x, y, rand.Intn(len(colors)))
		}
	}

	g.score = 0
	g.highlight = nil
	g.autoplay = false

	return g.ww, g.wh
}

func (g *Game) End() {
	w, h := g.blocks.Width(), g.blocks.Height()

	bw := (w - gow) / 2
	bh := (h - goh) / 2

	p := 0

	for y := 0; y < goh; y++ {
		for x := 0; x < gow; x++ {
			c := rand.Intn(len(colors)) // bg
			if gomessage[p] == 0 {
				c = bg // rand.Intn(len(colors))
			}

			g.blocks.Set(bw+x, bh+y, c)
			p++
		}
	}

}

func (g *Game) Print() {
	fmt.Println("[")
	for y := g.blocks.Height() - 1; y >= 0; y-- {
		fmt.Println(g.blocks.Row(y))
	}
	fmt.Println("]")
}

func (g *Game) Score() string {
	return fmt.Sprintf("%v", g.score)
}

func (g *Game) Coords(x, y int) (int, int) {
	return x / g.tw, g.blocks.Fix(y / g.th)
}

func (g *Game) ScreenCoords(x, y int) (int, int) {
	return x * g.tw, g.blocks.Fix(y) * g.th
}

func (g *Game) Connected(x, y int) []Point {
	v := g.blocks.Get(x, y)
	b := g.blocks.Clone()

	l, _ := connected(b, v, x, y, nil)

	w := b.Width()

	// sort top to bottom, right to left
	sort.SliceStable(l, func(i, j int) bool {
		w1 := l[i].y*w + (w - l[i].x)
		w2 := l[j].y*w + (w - l[j].x)

		return w2 < w1
	})

	return l
}

func connected(b matrix.Matrix[int], v, x, y int, list []Point) ([]Point, bool) {
	if x < 0 || x >= b.Width() || y < 0 || y >= b.Height() {
		return list, false
	}

	if b.Get(x, y) != v {
		return list, false
	}

	b.Set(x, y, visited)
	list = append(list, Point{x: x, y: y})
	list, _ = connected(b, v, x-1, y, list)
	list, _ = connected(b, v, x+1, y, list)
	list, _ = connected(b, v, x, y-1, list)
	list, _ = connected(b, v, x, y+1, list)
	return list, true
}

func (g *Game) Collapse(l []Point) {
	// first set to connected cells to empty
	for _, p := range l {
		g.blocks.Set(p.x, p.y, empty)
	}

	w, h := g.blocks.Width(), g.blocks.Height()

	// then collapse empty cells one column at a time
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if g.blocks.Get(x, y) == empty {
				for j := y + 1; j < h; j++ {
					g.blocks.Set(x, j-1, g.blocks.Get(x, j))
				}

				g.blocks.Set(x, h-1, bg)

				// I don't like this, but it works.
				// This is to cover the case where there are multiple empty cells in a column
				// (and I have been lazy and didn't want to optimize that case)
				y = -1
			}
		}
	}

	if len(l) >= nrefill {
		for y := h - 1; y >= 0; y-- {
			for x := 0; x < w; x++ {
				if g.blocks.Get(x, y) == bg {
					g.blocks.Set(x, y, rand.Intn(len(colors)))
				}
			}
		}
	}

	g.score += len(l)
}

func (g *Game) Find() []Point {
	for x := 0; x < hcount; x++ {
		for y := 0; y < vcount; y++ {
			if g.blocks.Get(x, y) == bg {
				continue
			}

			l := g.Connected(x, y)
			if len(l) >= nmatch {
				return l
			}
		}
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	tile := ebiten.NewImage(g.tw-border, g.th-border)

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			color := background

			ci := g.blocks.Get(x, y)
			switch {
			case ci >= 0:
				color = colors[ci]

			case ci == high:
				color = highlightColor
			}

			tile.Fill(color)

			sx, sy := g.ScreenCoords(x, y)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(sx+border), float64(sy+border))
			g.canvas.DrawImage(tile, op)
		}
	}

	tile.Fill(highlightColor)

	for _, p := range g.highlight {
		sx, sy := g.ScreenCoords(p.x, p.y)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx+border), float64(sy+border))
		g.canvas.DrawImage(tile, op)
	}

	screen.DrawImage(g.canvas, noop)
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyA): // (A)utoplay
		g.autoplay = !g.autoplay

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)estart
		g.Init(0, 0)
		ebiten.SetWindowTitle(title)

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyH): // (H)elp pressed
		if l := g.Find(); len(l) > 0 {
			g.highlight = l
		}

	case inpututil.IsKeyJustReleased(ebiten.KeyH): // (H)elp released
		g.highlight = nil

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		x, y := g.Coords(ebiten.CursorPosition())
		//fmt.Println(x, y)

		l := g.Connected(x, y)
		if len(l) < nmatch {
			break
		}

		g.Collapse(l)
		ebiten.SetWindowTitle(title + " - " + g.Score())

	case g.autoplay:
		if l := g.Find(); len(l) > 0 {
			g.Collapse(l)
			ebiten.SetWindowTitle(title + " - " + g.Score())
		} else {
			g.End()
			g.autoplay = false
		}
	}

	return nil
}
