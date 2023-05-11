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
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	bgColor        = color.NRGBA{80, 80, 80, 255}
	highlightColor = color.NRGBA{250, 250, 250, 255}

	colors = []color.NRGBA{
		{255, 0, 0, 255},     // red
		{0, 220, 0, 220},     // green
		{0, 0, 255, 255},     // blue
		{255, 220, 0, 255},   // yellow
		{64, 220, 220, 255},  // cyan
		{255, 128, 255, 255}, // magenta
	}

	noop = &ebiten.DrawImageOptions{}

	gomessage = []int{
		0, 1, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1,
		1, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 0, 0, 1, 1, 1, 1, 0,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1,
		0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1,
		1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0,
		1, 0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1,
		1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 0, 0,
		1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0,
		0, 1, 1, 1, 0, 0, 0, 1, 1, 1, 0, 0, 1, 0, 0, 0, 1, 0, 1, 1, 1, 1,
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

	ww, wh int // window width, height
	tw, th int // game tile width, height

	cx, cy int // last cell

	highlight []Point // blocks to hightlight
	autoplay  bool

	score int

	canvas *ebiten.Image // image buffer
	redraw bool          // content changed

	clear time.Time
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
	}

	for y := 0; y < vcount; y++ {
		for x := 0; x < hcount; x++ {
			g.blocks.Set(x, y, rand.Intn(len(colors)))
		}
	}

	g.score = 0
	g.highlight = nil
	g.autoplay = false
	g.redraw = true
	g.cx, g.cy = -1, -1

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
				c := rand.Intn(len(colors)) // bg
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

	g.score += 1 << len(l)
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
	if !g.clear.IsZero() && time.Now().After(g.clear) {
		g.clear = time.Time{}
	} else if !g.redraw {
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

	tile.Fill(highlightColor)

	for i, p := range g.highlight {
		if i == 0 {
			g.cx, g.cy = p.x, p.y
			g.clear = time.Now().Add(time.Second)
		}

		sx, sy := g.ScreenCoords(p.x, p.y)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx+border), float64(sy+border))
		g.canvas.DrawImage(tile, op)
	}

	if !(g.clear.IsZero()) {
		sx, sy := g.ScreenCoords(g.cx, g.cy)
		b := float32(border) / 2

		vector.StrokeRect(g.canvas,
			float32(sx)+b, float32(sy)+b, float32(g.tw), float32(g.th), b, highlightColor, false)
	}

	screen.DrawImage(g.canvas, noop)
	g.redraw = false
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyA): // (A)utoplay
		g.autoplay = !g.autoplay
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)estart
		g.Init(0, 0)
		ebiten.SetWindowTitle(title)
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyH): // (H)elp pressed
		if l := g.Find(); len(l) > 0 {
			g.highlight = l
			g.redraw = true
		}

	case inpututil.IsKeyJustReleased(ebiten.KeyH): // (H)elp released
		g.highlight = nil
		g.redraw = true

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		g.highlight = nil

		g.cx, g.cy = g.Coords(ebiten.CursorPosition())
		g.clear = time.Now().Add(time.Second)

		l := g.Connected(g.cx, g.cy)
		if len(l) < nmatch {
			break
		}

		g.redraw = true
		g.Collapse(l)
		ebiten.SetWindowTitle(title + " - " + g.Score())

		if l := g.Find(); len(l) == 0 {
			g.End()
		}

	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		l := g.Connected(g.cx, g.cy)
		if len(l) < nmatch {
			break
		}

		g.redraw = true
		g.Collapse(l)
		ebiten.SetWindowTitle(title + " - " + g.Score())

		if l := g.Find(); len(l) == 0 {
			g.End()
		}

	case isKeyPressed(ebiten.KeyLeft):
		g.redraw = true
		g.clear = time.Now().Add(time.Second)
		switch {
		case g.cx < 0:
			g.cx = g.blocks.Width() - 1
			g.cy = 0

		case g.cx == 0:
			g.cx = g.blocks.Width() - 1

		default:
			g.cx--
		}

	case isKeyPressed(ebiten.KeyRight):
		g.redraw = true
		g.clear = time.Now().Add(time.Second)
		switch {
		case g.cx < 0:
			g.cx = 0
			g.cy = 0

		case g.cx == g.blocks.Width()-1:
			g.cx = 0

		default:
			g.cx++
		}

	case isKeyPressed(ebiten.KeyDown):
		g.redraw = true
		g.clear = time.Now().Add(time.Second)
		switch {
		case g.cy < 0:
			g.cx = 0
			g.cy = 0

		case g.cy == 0:
			g.cy = g.blocks.Height() - 1

		default:
			g.cy--
		}

	case isKeyPressed(ebiten.KeyUp):
		g.redraw = true
		g.clear = time.Now().Add(time.Second)
		switch {
		case g.cy < 0:
			g.cx = 0
			g.cy = g.blocks.Height() - 1

		case g.cy == g.blocks.Height()-1:
			g.cy = 0

		default:
			g.cy++
		}

	case g.autoplay:
		g.redraw = true
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
