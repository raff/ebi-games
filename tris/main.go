package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"sort"
	"time"

	"math/rand"

	_ "embed"

	"github.com/disintegration/imaging"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	//"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	hcount = 10
	vcount = 4

	border = 8
)

var (
	//go:embed assets/tiles.png
	pngTiles []byte

	tiles []*ebiten.Image

	borderColor = color.NRGBA{80, 80, 80, 255}
	selectColor = color.NRGBA{255, 255, 0, 255}

	canvas *ebiten.Image

	tw, th int // game tile width, height
	gw, gh int // number of horizontal and vertical tiles in game
	ww, wh int // window width and height

	cards [][]int // gw columns of gh tiles (card indices)

	mcount = 2

	curmatches   = 0
	maxmatches   = 0
	totalmatches = 0
	shuffles     = 0

	mpoints = 1
	spoints = 1
	score   = 0
)

func initGame() {
	var lcards []int

	if len(tiles) == 0 {
		img, err := png.Decode(bytes.NewBuffer(pngTiles))
		if err != nil {
			log.Fatal(err)
		}

		isz := img.Bounds().Size()
		hsize := isz.X / hcount
		vsize := isz.Y / vcount

		tw = hsize / 2
		th = vsize / 2

		//card_loop:
		for v, y := 0, 0; v < vcount; v++ {
			for h, x := 0, 0; h < hcount; h++ {
				card := v*hcount + h

				tile := ebiten.NewImage(tw, th)
				tile.Fill(borderColor)
				im := imaging.Crop(img, image.Rect(x, y, x+hsize, y+vsize))
				im = imaging.Resize(im, tw-border, th-border-border, imaging.Box)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(border, border)
				tile.DrawImage(ebiten.NewImageFromImage(im), op)
				tiles = append(tiles, tile)

				lcards = append(lcards, card)
				lcards = append(lcards, card)
				lcards = append(lcards, card)
				lcards = append(lcards, card)
				lcards = append(lcards, card)
				lcards = append(lcards, card)

				x += hsize
			}

			y += vsize
		}

		gw, gh = factors(len(lcards))
		ww, wh = gw*tw, (gh+1)*th/2

		cards = make([][]int, gw)

		for i := range cards {
			cards[i] = make([]int, gh)
		}
	}

	cols := 0
	ci := -1

	if len(lcards) == 0 {
		for i, col := range cards {
			lcards = append(lcards, col...)

			if len(col) > 0 {
				cols++
				ci = i
			}
		}
	}

	if cols == 1 {
		// only one column left
		// make it one row

		cards[ci] = nil

		for i, c := range lcards {
			cards[i] = []int{c}
		}
	} else {
		rand.Shuffle(len(lcards), func(i, j int) {
			if lcards[i] != -1 && lcards[j] != -1 {
				lcards[i], lcards[j] = lcards[j], lcards[i]
			}
		})

		i := 0

		for x, col := range cards {
			for y := range col {
				cards[x][y] = lcards[i]
				i++
			}
		}
	}

	canvas = ebiten.NewImage(ww, wh)
	drawCards(nil)
}

func drawCards(revs map[int]bool) {
	canvas.Fill(borderColor)

	for x, col := range cards {
		for y, card := range col {
			im := tiles[card]
			ci := gameIndex(x, y)
			op := &ebiten.DrawImageOptions{}

			if revs != nil && revs[ci] {
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}

			op.GeoM.Translate(float64(x*tw), float64(y*th/2))
			canvas.DrawImage(im, op)
		}
	}
}

func cardIndex(x, y int) (int, int, int) {
	x /= tw
	y /= (th / 2)

	//log.Println("cardIndex", x, y)
	if x, y, c := playable(x, y); c >= 0 {
		return x, y, c
	}

	return -1, -1, -1
}

func playable(x, y int) (int, int, int) {
	if x >= 0 && x < gw && y >= 0 && y <= gh {
		col := cards[x]
		if y == len(col) { // bottom part of last card
			y--
		}
		if y == len(col)-1 { // last valid card in a column
			return x, y, col[y]
		}
	}

	return -1, -1, -1
}

func factors(n int) (int, int) {
	for i := n - 1; i > 1; i-- {
		if n%i == 0 {
			m1 := i
			m2 := n / i

			if m2 >= m1 {
				return m1, m2
			}
		}
	}

	return n, 1
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func getScore() string {
	return fmt.Sprintf("Tris - matches:%v  max.matches:%v  total:%v  shuffles:%v  score:%v  (%v)",
		curmatches, maxmatches, totalmatches, shuffles, score, spoints)
}

func main() {
	rand.Seed(time.Now().Unix())

	initGame()

	ebiten.SetWindowTitle("Tris")

	sw, sh := ebiten.ScreenSizeInFullscreen()
	w, h := ww/2, wh/2

	if sw/2 > w {
		w = sw / 2
	}

	if sh/2 > h {
		h = sh / 2
	}

	ebiten.SetWindowSize(w, h)

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}

	fmt.Println(getScore())
}

func gameIndex(x, y int) int {
	return y*gw + x
}

func gameCoord(gi int) (int, int) {
	y := gi / gw
	x := gi % gw
	return x, y
}

type Game struct{}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	//fmt.Println("layout", outsideWidth, outsideHeight, "-", ww, wh)
	return ww, wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(canvas, op)
}

var (
	matches  map[int]bool
	match    = -1
	autoplay = false
)

func (g *Game) Update() error {
	playCard := func(x, y int) error {
		x, y, card := playable(x, y)
		ci := gameIndex(x, y)

		if match != card {
			match = card
			matches = map[int]bool{ci: true}
		} else {
			matches[ci] = true
		}

		if len(matches) == mcount {
			for gi, _ := range matches {
				x, y := gameCoord(gi)
				cards[x] = cards[x][:y]
			}

			matches = nil
			match = -1

			curmatches++
			totalmatches++
			score += spoints
			if curmatches > maxmatches {
				maxmatches = curmatches
			}
			ebiten.SetWindowTitle(getScore())

			lc := 0

			for _, col := range cards {
				lc += len(col)
			}

			if lc == 0 {
				return fmt.Errorf("game over")
			}
		}

		drawCards(matches)
		return nil
	}

	shuffle := func() {
		initGame()

		shuffles++
		mpoints += curmatches
		spoints = int((float32(mpoints) + 0.5) / float32(shuffles))
		if spoints == 0 {
			spoints = 1
		}

		match = -1
		curmatches = 0
		ebiten.SetWindowTitle(getScore())
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyA):
		autoplay = !autoplay
		fmt.Println("autoplay:", autoplay)

	case autoplay:
		playcards := make([]struct{ x, y, c int }, len(cards))

		for x, col := range cards {
			y := len(col) - 1
			_, _, c := playable(x, y)
			playcards[x] = struct{ x, y, c int }{x, y, c}
		}

		sort.Slice(playcards, func(i, j int) bool {
			return playcards[i].c >= playcards[j].c
		})

		if playcards[0].c == -1 {
			fmt.Println("no valid cards")
			autoplay = false
			return fmt.Errorf("no valid cards")
		}

		matched := false

		for i := 0; i < len(playcards)-1; i++ {
			pc0 := playcards[i+0]
			pc1 := playcards[i+1]

			if pc0.c >= 0 && pc0.c == pc1.c {
				playCard(pc0.x, pc0.y)
				if err := playCard(pc1.x, pc1.y); err != nil {
					return err
				}

				matched = true
				break
			}
		}

		if !matched {
			shuffle()
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX):
		return fmt.Errorf("quit")

	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		shuffle()

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		x, y := ebiten.CursorPosition()
		if x, y, c := cardIndex(x, y); c >= 0 {
			if err := playCard(x, y); err != nil {
				return err
			}
		}
	}

	return nil
}
