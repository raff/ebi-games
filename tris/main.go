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

	mcount    = 3
	drawerLen = 7

	border = 8
)

type drawerCard struct {
	card int
	col  int
}

type drawerCards []drawerCard

type placeStatus int

const (
	cardPlaced placeStatus = iota
	cardMatched
	gameLost
)

var (
	//go:embed assets/tiles.png
	pngTiles []byte

	tiles []*ebiten.Image

	background  = color.NRGBA{80, 80, 80, 255}
	borderColor = color.NRGBA{127, 127, 127, 255}
	placeColor  = color.NRGBA{100, 100, 140, 255}

	canvas   *ebiten.Image
	drawerbg *ebiten.Image

	drawerOp = &ebiten.DrawImageOptions{}

	tw, th int // game tile width, height
	gw, gh int // number of horizontal and vertical tiles in game
	ww, wh int // window width and height

	cards  [][]int // gw columns of gh tiles (card indices)
	drawer drawerCards

	curmatches   = 0
	maxmatches   = 0
	totalmatches = 0
	shuffles     = 0

	mpoints = 1
	spoints = 1
	score   = 0

	autoplay = false

	selected = -1
)

func initGame(clear bool) {
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

		// add space for tiles drawer
		wh += border + border + th + border

		cards = make([][]int, gw)

		for i := range cards {
			cards[i] = make([]int, gh)
		}

		drawerbg = ebiten.NewImage(tw*drawerLen, th)
		drawerbg.Fill(placeColor)

		drawerOp.GeoM.Translate(float64(ww-drawerbg.Bounds().Dx())/2, float64(wh-th-border-border))

		drawer = make(drawerCards, 0, drawerLen)

		canvas = ebiten.NewImage(ww, wh)
	}

	cols := 0
	ci := -1

	if clear {
		for _, c := range drawer {
			cards[c.col] = append(cards[c.col], c.card)
		}

		drawer = drawer[:0]
	}

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

	drawCards(nil)
}

func placeCard(x, y, c int) placeStatus {
	drawer = append(drawer, drawerCard{card: c, col: x})

	sort.Slice(drawer, func(i, j int) bool {
		return drawer[i].card <= drawer[j].card
	})

	match := -1
	count := 0

	for i, pc := range drawer {
		if pc.card == match {
			count++

			if count == mcount {
				drawer = append(drawer[:i-mcount+1], drawer[i+1:]...)
				return cardMatched
			}
		} else {
			match = pc.card
			count = 1
		}
	}

	if len(drawer) == drawerLen {
		return gameLost
	}

	return cardPlaced
}

func drawCards(revs map[int]bool) {
	canvas.Fill(background)

	for x, col := range cards {
		for y, card := range col {
			im := tiles[card]
			ci := gameIndex(x, y)
			op := &ebiten.DrawImageOptions{}

			if revs != nil && revs[ci] {
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}
			if selected == x && y == len(col)-1 {
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}

			op.GeoM.Translate(float64(x*tw), float64(y*th/2))
			canvas.DrawImage(im, op)
		}
	}

	canvas.DrawImage(drawerbg, drawerOp)

	for i, p := range drawer {
		im := tiles[p.card]

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(((ww-drawerbg.Bounds().Dx())/2)+(i*tw)),
			float64(wh-th-border-border))

		canvas.DrawImage(im, op)
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

func nextCol(n int) int {
	for {
		selected += n
		if selected < 0 {
			selected = gw - 1
		} else if selected >= gw {
			selected = 0
		}

		if len(cards[selected]) > 0 {
			return selected
		}
	}

	return -1
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

	initGame(true)

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

func (g *Game) Update() error {
	playCard := func(x, y int) error {
		x, y, card := playable(x, y)
		cards[x] = cards[x][:y]

		status := placeCard(x, y, card)

		switch status {
		case cardMatched:
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

			if lc == 0 && len(drawer) == 0 {
				return fmt.Errorf("winner")
			}

		case gameLost:
			return fmt.Errorf("loser")
		}

		drawCards(nil)
		return nil
	}

	shuffle := func() {
		initGame(false)

		shuffles++
		mpoints += curmatches
		spoints = int((float32(mpoints) + 0.5) / float32(shuffles))
		if spoints == 0 {
			spoints = 1
		}

		curmatches = 0
		ebiten.SetWindowTitle(getScore())
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyA):
		autoplay = !autoplay
		fmt.Println("autoplay:", autoplay)

	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		nextCol(-1)
		drawCards(nil)

	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		nextCol(+1)
		drawCards(nil)

	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if selected >= 0 {
			x, y := selected, len(cards[selected])
			if y == 0 {
				selected = -1
			}

			if err := playCard(x, y); err != nil {
				return err
			}
		}

	case autoplay:
		selected = -1
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
		selected = -1
		shuffle()

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		x, y := ebiten.CursorPosition()
		if x, y, c := cardIndex(x, y); c >= 0 {
			if y == 0 {
				selected = -1
			}

			if err := playCard(x, y); err != nil {
				return err
			}
		}
	}

	return nil
}
