package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"log"
	"sort"
	"time"

	"math/rand"

	_ "embed"

	"github.com/raff/ebi-games/util"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	//"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	mcount    = 3
	drawerLen = 7

	border = 8
)

type drawerCard struct {
	card int
	col  int
}

type coords struct {
	x, y int
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

	tiles *util.Tiles

	background  = color.NRGBA{80, 80, 80, 255}
	borderColor = color.NRGBA{127, 127, 127, 255}
	placeColor  = color.NRGBA{100, 100, 140, 255}

	canvas   *ebiten.Image
	drawerbg *ebiten.Image

	drawerOp = &ebiten.DrawImageOptions{}

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

	level = 3
)

func initGame(clear bool) int {
	var lcards []int

	// load tiles from tile file
	if tiles == nil {
		var err error

		tiles, err = util.ReadTilesScaledTo(bytes.NewBuffer(pngTiles), 10, 4, 150, 150)
		if err != nil {
			log.Fatal(err)
		}

		for _, t := range tiles.List {
			b := t.Bounds()

			x := float32(b.Min.X + border)
			y := float32(b.Min.Y + border)
			w := float32(b.Dx() - border - border)
			h := float32(b.Dy() - border - border)

			vector.StrokeRect(t, x, y, w, h, float32(border), borderColor, false)
		}

	}

	// initialize cards
	if len(cards) == 0 {
		var copies, cardscount int

		switch level {
		case 1:
			cardscount = 10
			copies = 24
		case 2:
			cardscount = 20
			copies = 12
		case 3:
			cardscount = 26
			copies = 9
		case 4:
			cardscount = 40
			copies = 6
		}

		for card := 0; card < cardscount; card++ {
			for c := 0; c < copies; c++ {
				lcards = append(lcards, card)
			}
		}

		gw, gh = factors(len(lcards))
		ww, wh = gw*tiles.Width, (gh+1)*tiles.Height/2

		// add space for tiles drawer
		wh += border + border + tiles.Height + border

		cards = make([][]int, gw)

		for i := range cards {
			cards[i] = make([]int, gh)
		}
	}

	if canvas == nil {
		drawerbg = ebiten.NewImage(tiles.Width*drawerLen, tiles.Height)
		drawerbg.Fill(placeColor)

		drawerOp.GeoM.Translate(float64(ww-drawerbg.Bounds().Dx())/2, float64(wh-tiles.Width-border-border))

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
	return len(lcards)
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
			im := tiles.List[card]
			ci := gameIndex(x, y)
			op := &ebiten.DrawImageOptions{}

			if revs != nil && revs[ci] {
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}
			if selected == x && y == len(col)-1 {
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}

			op.GeoM.Translate(float64(x*tiles.Width), float64(y*tiles.Height/2))
			canvas.DrawImage(im, op)
		}
	}

	canvas.DrawImage(drawerbg, drawerOp)

	for i, p := range drawer {
		im := tiles.List[p.card]

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(((ww-drawerbg.Bounds().Dx())/2)+(i*tiles.Width)),
			float64(wh-tiles.Height-border-border))

		canvas.DrawImage(im, op)
	}

	ebiten.ScheduleFrame()
}

func cardIndex(x, y int) (int, int, int) {
	x /= tiles.Width
	y /= (tiles.Height / 2)

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
	return fmt.Sprintf("matches:%v  max.matches:%v  total:%v  shuffles:%v  score:%v  (%v)",
		curmatches, maxmatches, totalmatches, shuffles, score, spoints)
}

func main() {
	flag.IntVar(&level, "level", level, "difficulty level (1 to 4)")
	flag.Parse()

	rand.Seed(time.Now().Unix())

	if level < 1 {
		level = 1
	} else if level > 4 {
		level = 4
	}

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
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)

	err := ebiten.RunGame(&Game{})

	title := "Tris"
	if err != nil {
		title = err.Error()
	}

	fmt.Printf("%v / %v\n", title, getScore())
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
	playCard := func(x, y int, pause bool) error {
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
			ebiten.SetWindowTitle("Tris - " + getScore())

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
		if pause {
			time.Sleep(100 * time.Millisecond)
		}

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
		ebiten.SetWindowTitle("Tris - " + getScore())
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

			if err := playCard(x, y, false); err != nil {
				return err
			}
		}

	case autoplay:
		selected = -1

		matches := map[int]int{}        // card/count (cards in drawer)
		playables := map[int][]coords{} // card/list of card positions

		// get playable cards
		for x, col := range cards {
			y := len(col) - 1

			if y < 0 {
				continue
			}

			c := col[y]
			playables[c] = append(playables[c], coords{x, y})
		}

		// get cards in drawer
		for _, c := range drawer {
			matches[c.card]++
		}

		// if we can fit a full set and we do have full sets
		// play them
		if drawerLen-len(drawer) >= mcount {
			for _, cards := range playables {
				if len(cards) >= mcount {
					for i := 0; i < mcount; i++ {
						cc := cards[i]

						if err := playCard(cc.x, cc.y, true); err != nil {
							return err
						}
					}

					return nil
				}
			}
		}

		// if we can complete some matches in the drawer, do it
		for c, count := range matches {
			if count == mcount-1 && len(playables[c]) > 1 {
				cc := playables[c][0]

				return playCard(cc.x, cc.y, true)
			}
		}

		// if we can match some of the existing cards, play them
		if len(drawer) > 0 {
			for c, cards := range playables {
				if _, ok := matches[c]; ok {
					dl := len(drawer)    // cards in drawer
					dr := drawerLen - dl // remaining slots
					cl := len(cards)     // matching playable cards
					if cl > dr {
						cl = dr
					}

					if matches[c]+cl < mcount { // can't do a full match
						// make sure we don't fill the drawer
						if dr <= 1 {
							break
						}
					}

					for i := 0; i < cl; i++ {
						cc := cards[i]
						return playCard(cc.x, cc.y, true)
					}
				}
			}
		}

		// if there is space, pick a card and play it
		if len(drawer) < drawerLen-2 {
			for _, cards := range playables {
				if len(cards) > 0 {
					cc := cards[0]
					return playCard(cc.x, cc.y, true)
				}
			}
		}

		// no more matches
		shuffle()

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX):
		return fmt.Errorf("quit")

	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		curmatches = 0
		maxmatches = 0
		totalmatches = 0
		shuffles = 0
		mpoints = 1
		spoints = 1
		score = 0
		selected = -1

		cards = nil
		drawer = drawer[:0]
		initGame(true)
		ebiten.SetWindowTitle("Tris")

	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		selected = -1
		shuffle()

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		x, y := ebiten.CursorPosition()
		if x, y, c := cardIndex(x, y); c >= 0 {
			if y == 0 {
				selected = -1
			}

			if err := playCard(x, y, false); err != nil {
				return err
			}
		}
	}

	return nil
}
