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
	//"github.com/hajimehoshi/ebiten/v2/vector"
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

	title = "Colors"

	cRed    = 0
	cGreen  = 1
	cBlue   = 2
	cYellow = 3
	cOrange = 4
)

var (
	bgColor = color.NRGBA{80, 80, 80, 255}

	colors = []color.NRGBA{
		{255, 0, 0, 255},   // red
		{0, 220, 0, 220},   // green
		{0, 0, 255, 255},   // blue
		{255, 220, 0, 255}, // yellow
		{255, 125, 0, 255}, // orange
	}

	ncolors = len(colors)

	noop = &ebiten.DrawImageOptions{}

	gomessage = []int{
		0, 1, 1, 1, 0, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1,
		1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0,
		1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0,
		1, 0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0,
		0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 0,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0,
		1, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1,
		0, 1, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1,
	}

	gow = 22
	goh = 13
)

func main() {
	rand.Seed(time.Now().Unix())

	g := &Game{}

	ebiten.SetWindowTitle(title)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(ebiten.ScreenSizeInFullscreen()))
	ebiten.RunGame(g)
}

type Point struct {
	x, y int
}

type Game struct {
	blocks matrix.Matrix[int]
	ccount []int

	ww, wh int // window width, height
	tw, th int // game tile width, height

	score int
	size  int

	canvas *ebiten.Image // image buffer
	redraw bool          // content changed
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

		g.ww = (g.tw * hcount) + border
		g.wh = (g.th * vcount) + border

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(bgColor)

		g.blocks = matrix.New[int](hcount, vcount, false)
		g.ccount = make([]int, ncolors)
		g.size = len(g.blocks.Slice())
	}

	for i := 0; i < ncolors; i++ {
		g.ccount[i] = 0
	}

	g.blocks.Fill(bg)

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			c := rand.Intn(ncolors)
			g.blocks.Set(x, y, c)
			g.ccount[c]++
		}
	}

	g.score = 0
	g.redraw = true

	return g.ww, g.wh
}

func (g *Game) End() {
	g.blocks.Fill(bg)

	w, h := g.blocks.Width(), g.blocks.Height()

	bw := (w - gow) / 2
	bh := (h - goh) / 2

	p := 0

	for y := 0; y < goh; y++ {
		for x := 0; x < gow; x++ {
			if gomessage[p] != 0 {
				c := rand.Intn(ncolors) // bg
				g.blocks.Set(bw+x, bh+y, c)
			}

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

func (g *Game) Score(n int) string {
	return fmt.Sprintf("turn %v: %v/%v ", g.score, n, g.size)
}

func (g *Game) Coords(x, y int) (int, int) {
	return x / g.tw, g.blocks.Fix(y / g.th)
}

func (g *Game) ScreenCoords(x, y int) (int, int) {
	return x * g.tw, g.blocks.Fix(y) * g.th
}

func (g *Game) SetColor(c int) {
	pc := g.blocks.Get(0, 0)

	if pc == c {
		// already done
		return
	}

	l := g.Connected(0, 0)

	if len(l) == 0 {
		// nothing to do
		return
	}

	for _, p := range l {
		g.blocks.Set(p.x, p.y, c)
	}

	g.ccount[pc] -= len(l)
	g.ccount[c] += len(l)

	if g.ccount[c] == g.size {
		g.End()
	}

	g.score++
	g.redraw = true
	ebiten.SetWindowTitle(title + " - " + g.Score(g.ccount[c]))
}

func (g *Game) Connected(x, y int) []Point {
	v := g.blocks.Get(x, y)
	if v < 0 {
		return nil
	}

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

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	tile := ebiten.NewImage(g.tw-border, g.th-border)

	g.canvas.Fill(bgColor)

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			color := bgColor

			if ci := g.blocks.Get(x, y); ci >= 0 {
				color = colors[ci]
			}

			tile.Fill(color)

			sx, sy := g.ScreenCoords(x, y)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(sx+border), float64(sy+border))
			g.canvas.DrawImage(tile, op)
		}
	}

	screen.DrawImage(g.canvas, noop)
	g.redraw = false
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyN): // (N)ew game
		g.Init(0, 0)
		ebiten.SetWindowTitle(title)
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		cx, cy := g.Coords(ebiten.CursorPosition())
		g.SetColor(g.blocks.Get(cx, cy))

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // Red
		g.SetColor(cRed)

	case inpututil.IsKeyJustPressed(ebiten.KeyG): // Green
		g.SetColor(cGreen)

	case inpututil.IsKeyJustPressed(ebiten.KeyB): // Blue
		g.SetColor(cBlue)

	case inpututil.IsKeyJustPressed(ebiten.KeyY): // Yellow
		g.SetColor(cYellow)

	case inpututil.IsKeyJustPressed(ebiten.KeyO): // Orange
		g.SetColor(cOrange)
	}

	return nil
}

var keyPressed = map[ebiten.Key]bool{}

func isKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 10
		interval = 3
	)

	if inpututil.IsKeyJustReleased(key) {
		keyPressed[key] = false
		return false
	}

	d := inpututil.KeyPressDuration(key)
	if d > 0 && !keyPressed[key] {
		keyPressed[key] = true
		return true
	}

	if d >= delay && (d-delay)%interval == 0 {
		return true
	}

	return false
}
