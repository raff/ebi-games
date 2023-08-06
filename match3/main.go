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

	mincount = 3
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

type sequence struct {
	value, start, count int
}

func findseq(in []int) (out []sequence) {
	var curr sequence

	for i, v := range in {
		if v == curr.value {
			curr.count++
			continue
		}

		if curr.count >= mincount {
			out = append(out, curr)
		}

		curr = sequence{value: v, start: i, count: 1}
	}

	if curr.count >= mincount {
		out = append(out, curr)
	}

	return
}

type Game struct {
	cells matrix.Matrix[int]

	redraw  bool
	matches bool
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
	g.matches = true
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
	l := g.cells.Column(col)

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

func (g *Game) FindMatches() bool {
	w, h := g.cells.Width(), g.cells.Height()

	matched := false
	hseq := map[int][]sequence{}

	// check horizontal
	for i := 0; i < h; i++ {
		row := g.cells.Row(i)
		seq := findseq(row)

		if len(seq) == 0 {
			continue
		}

		//
		// cannot to this now, otherwise it messes up with columns
		//
		//for _, s := range seq {
		//	for x := 0; x < s.count; x++ {
		//		g.cells.Set(s.start+x, i, 0)
		//	}
		//}

		hseq[i] = seq // store the sequence and process later
		matched = true

	}

	// check vertical
	for i := 0; i < w; i++ {
		col := g.cells.Column(i)
		seq := findseq(col)

		if len(seq) == 0 {
			continue
		}

		matched = true

		for _, s := range seq {
			for y := 0; y < s.count; y++ {
				g.cells.Set(i, s.start+y, 0)
			}
		}
	}

	// update horizontal
	for i, seq := range hseq {
		for _, s := range seq {
			for x := 0; x < s.count; x++ {
				g.cells.Set(s.start+x, i, 0)
			}
		}
	}

	if matched {
		// collapse
		for i := 0; i < w; i++ {
			g.Collapse(i)
		}
	}

	return matched
}

func (g *Game) Update() error {
	if g.matches && g.FindMatches() {
		g.redraw = true
		return nil
	}

	g.matches = false

	checkMatch := func() {
		x, y := g.CellCoords(ebiten.CursorPosition())
		if x < 0 {
			return
		}

		g.redraw = true

		if g.highlight != nil {
			cells := g.cells.VonNewmann(x, y, false)
			for _, c := range cells {
				if c.X == g.highlight.X && c.Y == g.highlight.Y {
					g.cells.Swap(x, y, c.X, c.Y)

					g.matches = g.FindMatches()
					if !g.matches {
						g.cells.Swap(x, y, c.X, c.Y)
					}

					break
				}
			}

			g.highlight = nil
		} else {
			g.highlight = &image.Point{X: x, Y: y}
		}
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)edraw
		g.Init(0, 0)

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight): // Cycle colors
		if g.done {
			break
		}

		x, y := g.CellCoords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		v := g.cells.Get(x, y) + 1
		if v >= len(colors) {
			v = 1
		}

		g.cells.Set(x, y, v)
		g.redraw = true

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		if g.done {
			break
		}

		checkMatch()

	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft): // Mouse click and move
		if g.highlight != nil {
			x, y := g.CellCoords(ebiten.CursorPosition())
			if x < 0 {
				break
			}

			if x != g.highlight.X || y != g.highlight.Y {
				checkMatch()
			}
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
