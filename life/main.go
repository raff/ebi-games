package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	//"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	border = 1    // space between cells
	wrap   = true // wrap around
)

var (
	cwidth = 4 // cell width

	bgColor   = color.NRGBA{80, 80, 80, 255}
	cellColor = color.NRGBA{250, 250, 250, 255}

	noop = &ebiten.DrawImageOptions{}
)

func main() {
	wsize := flag.Int("window", 1, "Window size (1-4)")
	flag.IntVar(&cwidth, "cell", cwidth, "Cell size")
	start := flag.Int("start", 10, "Percentage of live cells at start")
	rule := flag.Int("rule", 1, "Rule: 1=life 2=highlife 3=34life 4=maze 5=mazectric 6=move")
	flag.Parse()

	rand.Seed(time.Now().Unix())

	if *start < 1 {
		*start = 1
	} else if *start > 100 {
		*start = 100
	}
	if *rule < 1 {
		*rule = 1
	} else if *rule > 6 {
		*rule = 6
	}

	g := &Game{rule: rules[*rule], start: *start}

	sw, sh := ebiten.ScreenSizeInFullscreen()

	switch *wsize {
	case 1: // half screen
		sw /= 2
		sh /= 2

	case 2: // 3/4 screen
		sw = sw * 3 / 4
		sh = sh * 3 / 4

	case 3:
		sw = sw * 7 / 8
		sh = sh * 7 / 8
	}

	//ebiten.SetWindowTitle(title)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(sw, sh))
	ebiten.RunGame(g)
}

type Rule func(alive bool, live int) bool

type ruleEntry struct {
	title string
	rule  Rule
}

var rules = []struct {
	title string
	rule  Rule
}{
	{"", nil},

	// Conway's Life
	//
	// Condesed rules:
	//  1. Any live cell with two or three live neighbours survives.
	//  2. Any dead cell with three live neighbours becomes a live cell.
	//  3. All other live cells die in the next generation.
	//     Similarly, all other dead cells stay dead.
	{"Game of Life", func(alive bool, live int) bool {
		if alive {
			return live == 2 || live == 3
		} else { // dead
			return live == 3
		}

		return false
	}},

	// 3-4 Life
	{"3-4 Life", func(alive bool, live int) bool {
		if alive {
			return live == 3 || live == 4
		} else { // dead
			return live == 3 || live == 4
		}

		return false
	}},

	// Highlife
	//
	// Condesed rules:
	//  1. Any live cell with two or three live neighbours survives.
	//  2. Any dead cell with three or siz live neighbours becomes a live cell.
	//  4. All other live cells die in the next generation.
	//     Similarly, all other dead cells stay dead.
	{"Highlife", func(alive bool, live int) bool {
		if alive {
			return live == 2 || live == 3
		} else { // dead
			return live == 3 || live == 6
		}

		return false
	}},

	// maze
	{"Maze", func(alive bool, live int) bool {
		if alive {
			return live >= 1 && live <= 5
		} else { // dead
			return live == 3
		}

		return false
	}},

	// mazectric
	{"Mazectric", func(alive bool, live int) bool {
		if alive {
			return live >= 1 && live <= 4
		} else { // dead
			return live == 3
		}

		return false
	}},

	// move
	{"Move", func(alive bool, live int) bool {
		if alive {
			return live == 2 || live == 4 || live == 5
		} else { // dead
			return live == 3 || live == 6 || live == 8
		}

		return false
	}},
}

type Game struct {
	world matrix.Matrix[bool]

	ww, wh int // window width, height
	tw, th int // game tile width, height

	canvas *ebiten.Image // image buffer
	cell   *ebiten.Image // cell image
	redraw bool          // content changed

	start int // % of live cells at gen 0
	rule  ruleEntry

	maxspeed int
	speed    int
	frame    int
	gen      int
}

func (g *Game) Init(w, h int) (int, int) {
	if w > 0 && h > 0 {
		g.ww, g.wh = w, h

		hcount := g.ww / cwidth
		vcount := g.wh / cwidth

		g.tw = g.ww / hcount
		g.th = g.wh / vcount

		g.ww = (g.tw * hcount) + border
		g.wh = (g.th * vcount) + border

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(bgColor)

		g.cell = ebiten.NewImage(g.tw-border, g.th-border)
		g.cell.Fill(cellColor)

		g.world = matrix.New[bool](hcount, vcount, false)

		g.speed = 2
		g.frame = g.speed
		g.maxspeed = 16
	} else {
		g.world.Fill(false)
	}

	w = g.world.Width()
	h = g.world.Height()

	for i := 0; i < len(g.world.Slice())*g.start/100; i++ {
		x := rand.Intn(w)
		y := rand.Intn(h)
		g.world.Set(x, y, true)
	}

	g.gen = 0
	return g.ww, g.wh
}

func (g *Game) End() {
}

func (g *Game) Print() {
	fmt.Println("[")
	for y := g.world.Height() - 1; y >= 0; y-- {
		fmt.Println(g.world.Row(y))
	}
	fmt.Println("]")
}

func (g *Game) Score() string {
	if g.speed == 0 {
		return fmt.Sprintf("<paused> generation: %v", g.gen)
	}

	return fmt.Sprintf("speed: %v generation: %v", g.speed, g.gen)
}

func (g *Game) Coords(x, y int) (int, int) {
	return x / g.tw, g.world.Fix(y / g.th)
}

func (g *Game) ScreenCoords(x, y int) (int, int) {
	return x * g.tw, g.world.Fix(y) * g.th
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	g.canvas.Fill(bgColor)

	for y := 0; y < g.world.Height(); y++ {
		for x := 0; x < g.world.Width(); x++ {
			if g.world.Get(x, y) {
				sx, sy := g.ScreenCoords(x, y)

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(sx+border), float64(sy+border))
				g.canvas.DrawImage(g.cell, op)
			}
		}
	}

	screen.DrawImage(g.canvas, noop)
	g.redraw = false

	ebiten.SetWindowTitle(g.rule.title + " - " + g.Score())
}

func (g *Game) Update() error {
	switch {
	case inpututil.MouseButtonPressDuration(ebiten.MouseButtonLeft) > 3: // Mouse click
		x, y := g.Coords(ebiten.CursorPosition())
		g.world.Set(x, y, !g.world.Get(x, y))
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)estart
		g.Init(0, 0)
		if g.speed == 0 {
			g.redraw = true
			g.frame = 1
		}

	case inpututil.IsKeyJustPressed(ebiten.Key1):
		g.redraw = true
		g.frame = 1
		g.rule = rules[1]

	case inpututil.IsKeyJustPressed(ebiten.Key2):
		g.redraw = true
		g.frame = 1
		g.rule = rules[2]

	case inpututil.IsKeyJustPressed(ebiten.Key3):
		g.redraw = true
		g.frame = 1
		g.rule = rules[3]

	case inpututil.IsKeyJustPressed(ebiten.Key4):
		g.redraw = true
		g.frame = 1
		g.rule = rules[4]

	case inpututil.IsKeyJustPressed(ebiten.Key5):
		g.redraw = true
		g.frame = 1
		g.rule = rules[5]

	case inpututil.IsKeyJustPressed(ebiten.Key6):
		g.redraw = true
		g.frame = 1
		g.rule = rules[6]

	case inpututil.IsKeyJustPressed(ebiten.KeyDown):
		if g.speed > 0 {
			g.speed -= 1
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyUp):
		if g.speed < g.maxspeed {
			g.speed += 1
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		g.speed = 1

	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		g.speed = g.maxspeed
	}

	if g.frame < g.maxspeed {
		if g.frame > 0 || g.speed > 0 { // g.frame <= 0 pauses the game
			g.frame++
		}

		return nil
	}

	g.frame = g.speed

	nw := matrix.NewLike(g.world)
	changes := false

	for y := 0; y < g.world.Height(); y++ {
		for x := 0; x < g.world.Width(); x++ {
			var live int // live neighbours

			alive := g.world.Get(x, y)

			for _, c := range g.world.Adjacent(x, y, wrap) {
				if c.Value {
					live++
				}
			}

			if g.rule.rule(alive, live) {
				nw.Set(x, y, true)
				changes = true
			}
		}
	}

	if changes {
		g.world = nw
		g.gen++
		g.redraw = true
	}

	return nil
}
