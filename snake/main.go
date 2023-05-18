package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	hcount = 25
	vcount = 25

	title = "Snake"
	max   = 10
)

var (
	bgColor    = color.NRGBA{40, 40, 40, 255}
	snakeColor = color.NRGBA{0, 255, 0, 255} // green
	foodColor  = color.NRGBA{255, 0, 0, 255} // green

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

type Dir int

const (
	Nodir Dir = iota
	Up
	Down
	Left
	Right
)

type Point struct {
	x, y int
}

type Snake struct {
	cells []Point
}

func NewSnake(x, y int) *Snake {
	return &Snake{cells: []Point{Point{x: x, y: y}}}
}

func (s *Snake) Head() Point {
	l := len(s.cells)
	return s.cells[l-1]
}

func (s *Snake) Move(d Dir, food Point) bool {
	p := s.Head()

	switch d {
	case Up:
		p.y++

	case Down:
		p.y--

	case Left:
		p.x--

	case Right:
		p.x++
	}

	s.cells = append(s.cells, p)
	add := p.x == food.x && p.y == food.y
	if !add {
		s.cells = s.cells[1:]
	}

	return add
}

type Game struct {
	snake *Snake
	food  Point
	dir   Dir

	ww, wh int // window width, height
	tw, th int // game tile width, height

	score int

	canvas *ebiten.Image // image buffer
	redraw bool          // content changed

	delay int
}

func RandXY() (int, int) {
	return rand.Intn(hcount), rand.Intn(vcount)
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

		g.ww = (g.tw * hcount)
		g.wh = (g.th * vcount)

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(bgColor)
	}

	g.snake = NewSnake(RandXY())
	g.food.x, g.food.y = RandXY()

	g.score = 0
	g.dir = Nodir
	g.redraw = true
	g.delay = 0

	return g.ww, g.wh
}

func (g *Game) End() {
}

func (g *Game) Score() string {
	return fmt.Sprintf("%v", g.score)
}

func (g *Game) Fix(y int) int {
	return hcount - 1 - y
}

func (g *Game) Coords(x, y int) (int, int) {
	return x / g.tw, g.Fix(y / g.th)
}

func (g *Game) ScreenCoords(x, y int) (int, int) {
	return x * g.tw, g.Fix(y) * g.th
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	tile := ebiten.NewImage(g.tw, g.th)

	g.canvas.Fill(bgColor)

	tile.Fill(snakeColor)

	for _, p := range g.snake.cells {
		sx, sy := g.ScreenCoords(p.x, p.y)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx), float64(sy))
		g.canvas.DrawImage(tile, op)
	}

	tile.Fill(foodColor)

	sx, sy := g.ScreenCoords(g.food.x, g.food.y)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(sx), float64(sy))
	g.canvas.DrawImage(tile, op)

	screen.DrawImage(g.canvas, noop)
	g.redraw = false
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyA): // (A)utoplay

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)estart

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case isKeyPressed(ebiten.KeyLeft):
		g.dir = Left

	case isKeyPressed(ebiten.KeyRight):
		g.dir = Right

	case isKeyPressed(ebiten.KeyDown):
		g.dir = Down

	case isKeyPressed(ebiten.KeyUp):
		g.dir = Up
	}

	if g.delay < max {
		g.delay++
		return nil
	}

	g.delay = 0

	h := g.snake.Head()
	eat := false

	switch g.dir {
	case Up:
		if h.y >= vcount-1 {
			break
		}
		eat = g.snake.Move(Up, g.food)

	case Down:
		if h.y <= 0 {
			break
		}
		eat = g.snake.Move(Down, g.food)

	case Left:
		if h.x <= 0 {
			break
		}
		eat = g.snake.Move(Left, g.food)

	case Right:
		if h.x >= hcount-1 {
			break
		}
		eat = g.snake.Move(Right, g.food)
	}

	if eat {
		g.score++
		g.food.x, g.food.y = RandXY()
	}

	g.redraw = true
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
