package main

import (
	//"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	_ "embed"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	border = 4

	hsize = 20
	vsize = 20

	csize = 30
)

var (
	background     = color.NRGBA{92, 92, 92, 255}
	highlightColor = color.NRGBA{250, 250, 250, 255}

	colors = []color.NRGBA{
		{0, 0, 0, 0},
		{250, 0, 0, 255},
		{0, 250, 0, 255},
		{0, 0, 250, 255},
		{250, 250, 0, 255},
		{250, 0, 250, 255},
		{0, 250, 250, 255},
		{250, 120, 20, 255},
	}
)

type Game struct {
	cells matrix.Matrix[int]

	redraw  bool
	started bool
	done    bool

	highlight *image.Point

	ww int // window width
	wh int // window height

	drawOp ebiten.DrawImageOptions
}

func (g *Game) Init(w, h int) (int, int) {
	if g.cells.Width() == 0 {
		g.cells = matrix.New[int](hsize, vsize, false)
	}

	cells := g.cells.Slice()
	nc := len(colors) - 1

	for i := range cells {
		v := rand.Intn(nc)
		cells[i] = v + 1
	}

	g.redraw = true
	g.started = false
	g.done = false

	g.ww, g.wh = hsize*csize+2*border, hsize*csize+2*border
	return g.ww, g.wh
}

func (g *Game) CellCoords(x, y int) (int, int) {
	if x < border || y < border {
		return -1, -1
	}
	if x > g.ww-border || y > g.wh-border {
		return -1, -1
	}

	return (x - border) / csize, g.cells.Fix((y - border) / csize)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	screen.Fill(background)

	cs := csize / 2
	r := float32(cs - 3)

	for y := 0; y < g.cells.Height(); y++ {
		cy := float32(border + y*csize + cs)

		for x := 0; x < g.cells.Width(); x++ {
			cx := float32(border + x*csize + cs)

			v := g.cells.Get(x, y)
			vector.DrawFilledCircle(screen, cx, cy, r, colors[v], true)

			if g.highlight != nil && g.highlight.X == x && g.highlight.Y == y {
				vector.StrokeCircle(screen, cx, cy, r, 2, highlightColor, true)
			}
		}
	}

	//g.highlight = nil
	g.redraw = false
}

func (g *Game) Collapse(col int) {
	l := g.cells.Col(col)

	for i := len(l) - 1; i >= 0; i-- {
		var j int

		for j = i; j >= 0; j-- {
			if l[j] != 0 {
				break
			}
		}

		if j == i {
			continue
		}
		if j == 0 && l[0] == 0 {
			break
		}

		for k := i; j >= 0; j-- {
			l[k], l[j] = l[j], 0
			k--
		}
	}

	nc := len(colors) - 1

	for i, v := range l {
		if v == 0 {
			v = rand.Intn(nc) + 1
		}
		g.cells.Set(col, i, v)
	}
}

func (g *Game) Matches(x, y int) (match bool) {
	v := g.cells.Get(x, y)
	if v == 0 {
		return false
	}

	row := g.cells.Row(y)
	minx, maxx := x, x

	for i := x - 1; i >= 0; i-- {
		if row[i] != v {
			break
		}

		minx = i
	}

	for i := x + 1; i <= len(row)-1; i++ {
		if row[i] != v {
			break
		}

		maxx = i
	}

	col := g.cells.Col(x)
	miny, maxy := y, y

	for i := y - 1; i >= 0; i-- {
		if col[i] != v {
			break
		}

		miny = i
	}

	for i := y + 1; i <= len(col)-1; i++ {
		if col[i] != v {
			break
		}

		maxy = i
	}

	if maxx-minx+1 >= 3 {
		match = true

		for i := minx; i <= maxx; i++ {
			g.cells.Set(i, y, 0)
		}
	}

	if maxy-miny+1 >= 3 {
		match = true

		for i := miny; i <= maxy; i++ {
			g.cells.Set(x, i, 0)
		}
	}

	if match {
		for i := 0; i < g.cells.Width(); i++ {
			g.Collapse(i)
		}
	}

	return
}

func (g *Game) AllMatches(x, y int) bool {
	matched := g.Matches(x, y)

	for {
		count := 0

		for y = 0; y < g.cells.Height(); y++ {
			for x = 0; x < g.cells.Width(); x++ {
				if g.Matches(x, y) {
					count++
					matched = true
				}
			}
		}

		if count == 0 {
			break
		}
	}

	return matched
}

func (g *Game) Update() error {
	if !g.started {
		g.started = true
		g.redraw = g.AllMatches(0, 0)
		return nil
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)edraw
		g.Init(0, 0)

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		if g.done {
			break
		}

		x, y := g.CellCoords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		if g.highlight != nil {
			cells := g.cells.VonNewmann(x, y, false)
			for _, c := range cells {
				if c.X == g.highlight.X && c.Y == g.highlight.Y {
					v1 := g.cells.Get(x, y)
					v2 := g.cells.Get(c.X, c.Y)
					g.cells.Set(x, y, v2)
					g.cells.Set(c.X, c.Y, v1)

					if !g.AllMatches(x, y) {
						g.cells.Set(x, y, v1)
						g.cells.Set(c.X, c.Y, v2)
					}

					break
				}
			}

			g.redraw = true
			g.highlight = nil
		} else {
			g.highlight = &image.Point{X: x, Y: y}
			g.redraw = true
		}
	}

	return nil
}

func main() {
	rand.Seed(time.Now().Unix())

	g := &Game{}
	ww, wh := ebiten.ScreenSizeInFullscreen()

	ebiten.SetWindowTitle("Match 3")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(ww, wh))
	ebiten.RunGame(g)
}
