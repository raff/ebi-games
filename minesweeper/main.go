package main

import (
	"bytes"
	"flag"
	//"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	_ "embed"

	"github.com/gobs/matrix"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/raff/ebi-games/util"
)

const (
	border = 4
)

type State int
type PlayState int

const (
	Unchecked State = iota
	Empty
	Flag
	Unsure
	UnsureChecked
	Mine
	Exploded
	Nomine

	Count1
	Count2
	Count3
	Count4
	Count5
	Count6
	Count7
	Count8

	MineFlag
	MineUnsure
	MineUnsureChecked
)

const (
	Playing PlayState = iota
	Clicked
	Surprise
	Won
	Lost
)

type Level struct {
	width  int
	height int
	mines  int
}

var (
	//go:embed assets/ms_cells.png
	pngCells []byte

	//go:embed assets/ms_counts.png
	pngCounts []byte

	//go:embed assets/ms_digits.png
	pngDigits []byte

	//go:embed assets/ms_faces.png
	pngFaces []byte

	tiles  *util.Tiles
	digits *util.Tiles
	faces  *util.Tiles

	background = color.NRGBA{192, 192, 192, 255}

	levels = []Level{
		{9, 9, 10},
		{16, 16, 40},
		{30, 16, 99},
	}
)

type Game struct {
	level Level

	cells matrix.Matrix[State]

	canvas *ebiten.Image
	redraw bool
	done   bool

	state PlayState

	start   time.Time
	elapsed int

	mx int // mine count area start (x)
	my int // mine count area start (y)

	tx int // timer area start (x)
	ty int // timer area start (y)

	fx int // face area start (x)
	fy int // face area start (y)

	cw int // cells area width
	ch int // cells area height

	cx int // cells area start (x)
	cy int // cells area start (y)

	ww int // window width
	wh int // window height

	scale float64

	drawOp ebiten.DrawImageOptions
}

func (g *Game) Init(w, h int, scale float64) (int, int) {
	if tiles == nil {
		var err error

		tiles, err = util.ReadTiles(bytes.NewBuffer(pngCells), 8, 1)
		if err != nil {
			log.Fatal(err)
		}

		tt, err := util.ReadTiles(bytes.NewBuffer(pngCounts), 8, 1)
		if err != nil {
			log.Fatal(err)
		}

		tiles.List = append(tiles.List, tt.List...)
		tiles.List = append(tiles.List, tiles.List[Flag])          // MineFlag
		tiles.List = append(tiles.List, tiles.List[Unsure])        // MineUnsure
		tiles.List = append(tiles.List, tiles.List[UnsureChecked]) // MineUnsureChecked

		digits, err = util.ReadTiles(bytes.NewBuffer(pngDigits), 10, 1)
		if err != nil {
			log.Fatal(err)
		}

		faces, err = util.ReadTiles(bytes.NewBuffer(pngFaces), 5, 1)
		if err != nil {
			log.Fatal(err)
		}
	}

	if g.ww == 0 {
		g.cw = g.level.width * tiles.Width
		g.ch = g.level.height * tiles.Height

		g.cx = border
		g.cy = border + digits.Height + border

		g.ww = g.cx + g.cw + border
		g.wh = g.cy + g.ch + border

		g.mx = border
		g.my = border

		g.tx = g.ww - border - (digits.Width * 3) // 3 digits timer
		g.ty = border

		g.fx = (g.ww / 2) - (faces.Width / 2)
		g.fy = border

		g.canvas = ebiten.NewImage(g.ww, g.wh)
		g.canvas.Fill(background)

		g.cells = matrix.New[State](g.level.width, g.level.height, false)

		g.scale = scale

		if g.scale != 1.0 {
			g.drawOp.GeoM.Scale(g.scale, g.scale)
			g.ww = int(float64(g.ww) * g.scale)
			g.wh = int(float64(g.wh) * g.scale)
		}
	}

	states := g.cells.Slice()
	perms := rand.Perm(len(states))

	for i, p := range perms {
		if p < g.level.mines {
			states[i] = Mine
		} else {
			states[i] = Unchecked
		}
	}

	g.start = time.Time{}
	g.elapsed = 0

	g.state = Playing
	g.redraw = true
	g.done = false
	return g.ww, g.wh
}

func (g *Game) FaceClicked(x, y int) bool {
	x = int(float64(x) / g.scale)
	y = int(float64(y) / g.scale)

	if x < g.fx || y < g.fy {
		return false
	}

	if x >= g.fx+faces.Width || y >= g.fy+faces.Height {
		return false
	}

	return true
}

func (g *Game) CellCoords(x, y int) (int, int) {
	x = int(float64(x) / g.scale)
	y = int(float64(y) / g.scale)

	if x < g.cx || y < g.cy {
		return -1, -1
	}
	if x > g.cx+g.cw || y > g.cy+g.ch {
		return -1, -1
	}

	return (x - g.cx) / tiles.Width, g.cells.Fix((y - g.cy) / tiles.Height)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) drawDigits(x, y, n int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))

	for x := 100; x > 0; x /= 10 {
		d := (n / x) % 10

		g.canvas.DrawImage(digits.Item(d), op)
		op.GeoM.Translate(float64(digits.Width), 0)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.cx), float64(g.cy))

	found := 0
	wrong := 0

	for y := 0; y < g.level.height; y++ {
		for x := 0; x < g.level.width; x++ {
			s := g.cells.Get(x, y)
			if g.done {
				switch s {
				case Unchecked, Unsure:
					s = Empty
				case Flag:
					s = Nomine
					wrong++
				case MineFlag:
					found++
				}
			} else if s == Mine {
				s = Unchecked
			} else if s == Nomine {
				wrong++
			} else if s == Flag || s == MineFlag {
				found++
			}

			g.canvas.DrawImage(tiles.List[s], op)
			op.GeoM.Translate(float64(tiles.Width), 0)
		}

		op.GeoM.SetElement(0, 2, border)
		op.GeoM.Translate(0, float64(tiles.Height))
	}

	if !g.done && found == g.level.mines {
		if wrong > 0 {
			g.state = Lost
		} else {
			g.state = Won
		}

		g.done = true
		g.redraw = true
	}

	g.drawDigits(g.mx, g.my, found)
	g.drawDigits(g.tx, g.ty, g.elapsed)

	op.GeoM.Reset()
	op.GeoM.Translate(float64(g.fx), float64(g.fy))
	g.canvas.DrawImage(faces.List[g.state], op)

	screen.DrawImage(g.canvas, &g.drawOp)
}

func (g *Game) countMines(x, y int) (int, []matrix.Cell[State]) {
	count := 0
	cells := g.cells.Moore(x, y, false)

	for _, c := range cells {
		if c.Value == Mine || c.Value >= MineFlag {
			count++
		}
	}

	return count, cells
}

func (g *Game) expand(cells []matrix.Cell[State]) {
	for _, c := range cells {
		if c.Value != Unchecked {
			continue
		}

		cc, ccells := g.countMines(c.X, c.Y)
		if cc > 0 {
			s := Count1 + State(cc-1)
			g.cells.Set(c.X, c.Y, s)
		} else {
			g.cells.Set(c.X, c.Y, Empty)
			g.expand(ccells)
		}
	}
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX): // (Q)uit or e(X)it
		return ebiten.Termination

	case inpututil.IsKeyJustPressed(ebiten.KeyR): // (R)edraw
		g.Init(0, 0, 0)

	case inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft): // Mouse release
		if g.FaceClicked(ebiten.CursorPosition()) {
			g.Init(0, 0, 0)
			break
		}

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft): // Mouse click
		if g.FaceClicked(ebiten.CursorPosition()) {
			g.state = Clicked
			break
		}

		if g.done {
			break
		}

		x, y := g.CellCoords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		switch g.cells.Get(x, y) {
		case Unchecked:
			count, cells := g.countMines(x, y)

			if count > 0 {
				s := Count1 + State(count-1)
				g.cells.Set(x, y, s)
			} else {
				g.cells.Set(x, y, Empty)
				g.expand(cells)
			}

			g.redraw = true

		case Mine:
			g.cells.Set(x, y, Exploded)
			g.state = Lost
			g.redraw = true
			g.done = true
		}

		if g.start.IsZero() {
			g.start = time.Now()
		} else {
			g.elapsed = int(time.Now().Sub(g.start) / time.Second)
		}

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight): // Mouse click
		if g.done {
			break
		}

		x, y := g.CellCoords(ebiten.CursorPosition())
		if x < 0 {
			break
		}

		switch g.cells.Get(x, y) {
		case Unchecked:
			g.cells.Set(x, y, Flag)
			g.redraw = true

		case Flag:
			g.cells.Set(x, y, Unsure)
			g.redraw = true

		case Unsure:
			g.cells.Set(x, y, Unchecked)
			g.redraw = true

		case Mine:
			g.cells.Set(x, y, MineFlag)
			g.redraw = true

		case MineFlag:
			g.cells.Set(x, y, MineUnsure)
			g.redraw = true

		case MineUnsure:
			g.cells.Set(x, y, Mine)
			g.redraw = true
		}

		if g.start.IsZero() {
			g.start = time.Now()
		} else {
			g.elapsed = int(time.Now().Sub(g.start) / time.Second)
		}
	}

	return nil
}

func main() {
	scale := flag.Float64("scale", 2, "Window scale")
	level := flag.Int("level", 0, "0-beginner, 1-intermediat, 2-expert")
	flag.Parse()

	rand.Seed(time.Now().Unix())

	if *level < 0 {
		*level = 0
	} else if *level >= len(levels) {
		*level = len(levels) - 1
	}

	g := &Game{level: levels[*level]}
	ww, wh := ebiten.ScreenSizeInFullscreen()

	ebiten.SetWindowTitle("Minesweeper")
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.Init(ww, wh, *scale))
	ebiten.RunGame(g)
}
