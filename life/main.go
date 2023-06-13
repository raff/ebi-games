package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"time"

	_ "embed"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	border = 1    // space between cells
	wrap   = true // wrap around
)

var (
	//go:embed rules.txt
	rulesFile string

	cwidth = 4 // cell width

	bgColor   = color.NRGBA{80, 80, 80, 255}
	cellColor = color.NRGBA{250, 250, 250, 255}

	noop = &ebiten.DrawImageOptions{}

	rules = []Rule{ParseRule("Conway's Life:B3/S23")}
)

func readRules() {
	scanner := bufio.NewScanner(strings.NewReader(rulesFile))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if r := ParseRule(line); r.title != "" {
			rules = append(rules, r)
		}
	}
}

func main() {
	wsize := flag.Int("window", 1, "Window size (1-4)")
	flag.IntVar(&cwidth, "cell", cwidth, "Cell size")
	start := flag.Int("start", 10, "Percentage of live cells at start")
	rstring := flag.String("rulestring", "", "Rule string in the format `title:Bxxx/Sxxx`")
	mirror := flag.Bool("mirror", false, "mirror world to 4 quadrants (kaleidoscope)")
	color := flag.Bool("color", false, "use colors for cell age")
	flag.Parse()

	rand.Seed(time.Now().Unix())

	if *start < 1 {
		*start = 1
	} else if *start > 100 {
		*start = 100
	}

	readRules()

	if *rstring != "" {
		if r := ParseRule(*rstring); r.title != "" {
			rules[0] = r
		}
	}

	g := &Game{rule: 0, start: *start, mirror: *mirror, color: *color}

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

	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(sw, sh))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.RunGame(g)
}

type Rule struct {
	title string
	dead  int // bitmask of how many live cells are needed to make the cell alive
	live  int // bitmask of how many live cells are needed to keep the cell alive
	age   int // a live cell will age (+1) at each generation until it reaches this value (and dies)
}

const (
	NoState      = 0
	BirthState   = 1
	SurviveState = 2

	CellDead  = 0
	CellAlive = 1
	CellAging = 2
)

func ParseRule(s string) (r Rule) {
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		r.title = "No name"

	case 2:
		r.title = parts[0]
		s = parts[1]

	default:
		panic("invalid rule string")
	}

	state := NoState
	age := 0

	for _, c := range s {
		switch state {
		case NoState:
			switch c {
			case 'B', 'b':
				state = BirthState

			case 'S', 's':
				state = SurviveState

			default:
				if c >= '0' && c <= '9' {
					age = (age * 10) + int(c-'0')
				}
			}

		case BirthState:
			if c >= '0' && c <= '9' {
				r.dead |= 1 << (c - '0')
			} else {
				state = NoState
			}

		case SurviveState:
			if c >= '0' && c <= '9' {
				r.live |= 1 << (c - '0')
			} else {
				state = NoState
			}
		}
	}

	if age == 0 {
		r.age = 2
	} else {
		r.age = age
	}
	return r
}

func (r Rule) Check(age int, live int) int {
	bit := 1 << live

	switch age {
	case CellDead:
		if r.dead&bit == bit {
			age = CellAlive
		}

	case CellAlive:
		if r.live&bit != bit {
			age++
		}

	default:
		age++
	}

	return age % r.age
}

func bset(b int) int {
	return 1 << b
}

type Game struct {
	world matrix.Matrix[int]

	ww, wh int // window width, height
	tw, th int // game tile width, height

	canvas *ebiten.Image // image buffer
	cell   *ebiten.Image // cell image
	redraw bool          // content changed

	start  int  // % of live cells at gen 0
	rule   int  // index in rules list
	mirror bool // mirror life's world in to quadrants
	color  bool // use colors for cell age

	maxspeed int
	speed    int
	frame    int
	gen      int
}

func (g *Game) Init(w, h int) (int, int) {
	gen := true

	if w > 0 && h > 0 {
		g.ww, g.wh = w, h

		hcount := g.ww / cwidth
		vcount := g.wh / cwidth

		if g.mirror {
			// since we are dividing in half, make sure we have an even number of cells
			if hcount%2 != 0 {
				hcount--
			}
			if vcount%2 != 0 {
				vcount--
			}
		}

		g.tw = g.ww / hcount
		g.th = g.wh / vcount

		g.ww = (g.tw * hcount) + border
		g.wh = (g.th * vcount) + border

		if g.mirror {
			g.canvas = ebiten.NewImage((g.ww-border)/2, (g.wh-border)/2)
			hcount /= 2
			vcount /= 2
		} else {
			g.canvas = ebiten.NewImage(g.ww, g.wh)
		}

		g.canvas.Fill(bgColor)

		g.cell = ebiten.NewImage(g.tw-border, g.th-border)
		g.cell.Fill(cellColor)

		if g.maxspeed == 0 {
			g.maxspeed = 16
			g.speed = 2
			g.frame = g.speed
		}

		currw := g.world
		cw, ch := currw.Width(), currw.Height()

		g.world = matrix.New[int](hcount, vcount, false)

		if cw != 0 && ch != 0 {
			if hcount >= cw || vcount >= ch {
				x := (hcount - cw) / 2
				y := (vcount - ch) / 2

				g.world.Copy(x, y, currw)
			} else {
				x := (cw - hcount) / 2
				y := (ch - vcount) / 2

				g.world = currw.Submatrix(x, y, hcount, vcount)
			}

			gen = false
		}
	} else {
		g.world.Fill(CellDead)
	}

	if gen {
		w = g.world.Width()
		h = g.world.Height()

		for i := 0; i < len(g.world.Slice())*g.start/100; i++ {
			x := rand.Intn(w)
			y := rand.Intn(h)
			g.world.Set(x, y, CellAlive)
		}

		g.gen = 0
	}

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

func (g *Game) Details() string {
	title := fmt.Sprintf("%d: %v - ", g.rule, rules[g.rule].title)

	if g.speed == 0 {
		return title + fmt.Sprintf("<paused> generation: %v", g.gen)
	}

	return title + fmt.Sprintf("speed: %v generation: %v", g.speed, g.gen)
}

func (g *Game) Coords(x, y int) (int, int) {
	if x < 0 {
		x = 0
	}
	if x >= g.ww {
		x = g.ww - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= g.wh {
		y = g.wh - 1
	}

	if g.mirror {
		cw, ch := g.canvas.Size()

		if x >= cw {
			x = g.ww - border - x
		}
		if y >= ch {
			y = g.wh - border - y
		}
	}

	return x / g.tw, g.world.Fix(y / g.th)
}

func (g *Game) ScreenCoords(x, y int) (int, int) {
	return x * g.tw, g.world.Fix(y) * g.th
}

func abs(i int) int {
	if i < 0 {
		return -i
	}

	return i
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if abs(outsideWidth-g.ww) <= (cwidth+border) && abs(outsideHeight-g.wh) <= (cwidth+border) {
		return g.ww, g.wh
	}

	return g.Init(outsideWidth, outsideHeight)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	g.canvas.Fill(bgColor)

	rule := rules[g.rule]
	scale := float32(rule.age)

	rscale := rand.Int()
	scaleR := rscale&1 != 0
	scaleG := rscale&2 != 0
	scaleB := rscale&4 != 0

	for y := 0; y < g.world.Height(); y++ {
		for x := 0; x < g.world.Width(); x++ {
			if age := g.world.Get(x, y); age > CellDead {
				sx, sy := g.ScreenCoords(x, y)

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(sx+border), float64(sy+border))

				r := float32(age+1) / scale

				if g.color {
					if scaleR {
						op.ColorScale.SetR(r)
					}
					if scaleG {
						op.ColorScale.SetG(r)
					}
					if scaleB {
						op.ColorScale.SetB(r)
					}
				} else {
					op.ColorScale.ScaleAlpha(r)
				}

				g.canvas.DrawImage(g.cell, op)
			}
		}
	}

	screen.DrawImage(g.canvas, noop)
	if g.mirror {
		op := noop
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(g.ww-border), 0)
		screen.DrawImage(g.canvas, op)

		op.GeoM.Reset()
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, float64(g.wh-border))
		screen.DrawImage(g.canvas, op)

		op.GeoM.Reset()
		op.GeoM.Scale(-1, -1)
		op.GeoM.Translate(float64(g.ww-border), float64(g.wh-border))
		screen.DrawImage(g.canvas, op)

		op.GeoM.Reset()
	}

	g.redraw = false

	ebiten.SetWindowTitle(g.Details())
}

func (g *Game) Update() error {
	k := numberKey()

	switch {
	case inpututil.MouseButtonPressDuration(ebiten.MouseButtonLeft) > 3: // Mouse click
		x, y := g.Coords(ebiten.CursorPosition())
		c := g.world.Get(x, y)
		if c == CellDead {
			c = CellAlive
		} else {
			c = CellDead
		}
		g.world.Set(x, y, c)
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)estart
		g.Init(0, 0)
		if g.speed == 0 {
			g.redraw = true
			g.frame = 1
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft): // [
		g.rule--
		if g.rule < 0 {
			g.rule = len(rules) - 1
		}
		g.frame = 1
		g.redraw = true

	case inpututil.IsKeyJustPressed(ebiten.KeyBracketRight): // ]
		g.rule++
		if g.rule >= len(rules) {
			g.rule = 0
		}
		g.frame = 1
		g.redraw = true

	case k >= 0:
		if k < len(rules) {
			g.rule = k
			g.frame = 1
			g.redraw = true
		}

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
	rule := rules[g.rule]

	for y := 0; y < g.world.Height(); y++ {
		for x := 0; x < g.world.Width(); x++ {
			var live int // live neighbours

			age := g.world.Get(x, y)

			for _, c := range g.world.Adjacent(x, y, wrap) {
				if c.Value == CellAlive {
					live++
				}
			}

			if age = rule.Check(age, live); age != CellDead {
				nw.Set(x, y, age)
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

func numberKey() int {
	for _, k := range inpututil.PressedKeys() {
		if k >= ebiten.KeyDigit0 && k <= ebiten.KeyDigit9 {
			if ebiten.IsKeyPressed(ebiten.KeyControl) {
				return int(k-ebiten.KeyDigit0) + 10
			}

			return int(k - ebiten.KeyDigit0)
		}
	}

	return -1
}
