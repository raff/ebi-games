package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	cw = 14

	border = 6

	title = "Snake"
)

var (
	bgColor    = color.NRGBA{40, 40, 40, 255}
	snakeColor = color.NRGBA{0, 255, 0, 255} // green
	foodColor  = color.NRGBA{255, 0, 0, 255} // green

	noop = &ebiten.DrawImageOptions{}
)

func main() {
	rand.Seed(time.Now().Unix())

	g := &Game{}

	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(ebiten.ScreenSizeInFullscreen()))
	ebiten.SetWindowTitle(g.Score())
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

type Dir int

const (
	Nodir Dir = iota
	Up
	Down
	Left
	Right
)

type MoveResult int

const (
	Space MoveResult = iota
	Wall
	Body
	Food
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

func (s *Snake) Move(d Dir, w, h int, food Point) MoveResult {
	p := s.Head()

	switch d {
	case Up:
		if p.y >= h-1 {
			return Wall
		}
		p.y++

	case Down:
		if p.y <= 0 {
			return Wall
		}
		p.y--

	case Left:
		if p.x <= 0 {
			return Wall
		}
		p.x--

	case Right:
		if p.x >= w-1 {
			return Wall
		}
		p.x++
	}

        for _, b := range s.cells {
            if p.x == b.x && p.y == b.y {
                return Body
            }
        }

	s.cells = append(s.cells, p)
	if p.x == food.x && p.y == food.y {
		return Food
	}

	s.cells = s.cells[1:]
	return Space
}

type Game struct {
	snake *Snake
	food  Point
	dir   Dir

	message string

	ww, wh int // window width, height
	tw, th int // game tile width, height

	cols, rows int

	score int
	lives int

	canvas *ebiten.Image // image buffer
	redraw bool          // content changed

	frame    int
	speed    int
	maxspeed int

	starve int
	eats   int
}

func (g *Game) RandXY() (int, int) {

retry:
	for {
		x, y := rand.Intn(g.cols), rand.Intn(g.rows)

		if g.food.x == x && g.food.y == y {
			continue retry
		}

		if g.snake != nil {
			for _, p := range g.snake.cells {
				if p.x == x && p.y == y {
					continue retry
				}
			}
		}

		return x, y
	}
}

/*
 * game.Init(w, h) : new game, calculate all dimensions
 *
 * game.Init(0, 0) : restart, reset all values
 *
 * game.Init(-1, -1) : new life
 *
 */
func (g *Game) Init(w, h int) (int, int) {
	if w > 0 && h > 0 {
		g.ww, g.wh = w/2, h/2

		g.tw = cw // g.ww / hcount
		g.th = cw // g.wh / vcount

		g.cols = g.ww / g.tw
		g.rows = g.wh / g.th

		g.ww = (g.tw * g.cols) + (2 * border)
		g.wh = (g.th * g.rows) + (2 * border)

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(bgColor)
	}

	g.snake = NewSnake(g.RandXY())
	g.food.x, g.food.y = g.RandXY()

	g.dir = Nodir
	g.redraw = true
	g.speed = 1
	g.frame = g.speed
	g.maxspeed = 10
	g.starve = 0
	g.eats = 5
	g.message = ""

	if w >= 0 {
		g.score = 0
		g.lives = 5
	}

	return g.ww, g.wh
}

func (g *Game) End() {
}

func (g *Game) Score() string {
	return fmt.Sprintf("%v - score: %v speed: %v lives: %v", title, g.score, g.speed, g.lives)
}

func (g *Game) Fix(y int) int {
	return g.rows - 1 - y
}

func (g *Game) ScreenCoords(x, y int) (int, int) {
	return border + (x * g.tw), border + (g.Fix(y) * g.th)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}
	if g.message != "" {
		screen.Fill(bgColor)
		ebitenutil.DebugPrint(screen, g.message)
		g.redraw = false
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
	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)estart
		g.Init(0, 0)
		ebiten.SetWindowTitle(g.Score())
		return nil

	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if g.frame > 0 {
			g.frame = 0
		} else if g.message != "" {
			if g.lives > 0 { // new life
				g.Init(-1, -1)
		                ebiten.SetWindowTitle(g.Score())
			}
			// else game over: restart
		} else {
			g.frame = g.speed
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		g.dir = Left

	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		g.dir = Right

	case inpututil.IsKeyJustPressed(ebiten.KeyDown):
		g.dir = Down

	case inpututil.IsKeyJustPressed(ebiten.KeyUp):
		g.dir = Up
	}

	if g.frame < g.maxspeed {
		if g.frame > 0 { // g.frame <= 0 pauses the game
			g.frame++
		}

		return nil
	}

	g.frame = g.speed

	var mres MoveResult

	switch g.dir {
	case Up:
		mres = g.snake.Move(Up, g.cols, g.rows, g.food)

	case Down:
		mres = g.snake.Move(Down, g.cols, g.rows, g.food)

	case Left:
		mres = g.snake.Move(Left, g.cols, g.rows, g.food)

	case Right:
		mres = g.snake.Move(Right, g.cols, g.rows, g.food)
	}

	switch mres {
	case Wall, Body:
		g.lives--
		if g.lives > 0 {
			g.frame = 0
			g.message = " *** CRASH - Hit <space> to continue ***"
		} else {
			g.message = " *** GAME OVER ***"
		}

		g.dir = Nodir
		ebiten.SetWindowTitle(g.Score())

	case Food:
		g.score++
		g.starve++
		if g.starve >= g.eats && g.speed < g.maxspeed {
			g.speed++
			g.starve = 0
			g.eats += 2
		}

		g.food.x, g.food.y = g.RandXY()
		ebiten.SetWindowTitle(g.Score())
	}

	g.redraw = true
	return nil
}
