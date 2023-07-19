package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"log"
	"time"

	"math/rand"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/raff/ebi-games/util"
)

const (
	hcount = 10
	vcount = 4
)

var (
	//go:embed assets/tiles.png
	pngTiles []byte

	tiles *util.Tiles

	cover       *ebiten.Image
	coverColor  = color.NRGBA{0, 0, 32, 255}
	borderColor = color.NRGBA{200, 200, 255, 0}

	waitTurn = 300 * time.Millisecond
	waitGame = 5 * time.Second

	canvas *ebiten.Image

	gw, gh int // number of horizontal and vertical tiles in game

	maxcards = hcount * vcount

	cards  []int  // gw * gh tiles, card indices
	states []bool // gw * gh tiles, card states

	moves   int
	matches int
)

func initGame() {
	if tiles == nil {
		var err error

		tiles, err = util.ReadTilesScaledTo(bytes.NewBuffer(pngTiles), 10, 4, 150, 150)
		if err != nil {
			log.Fatal(err)
		}

	card_loop:
		for v := 0; v < tiles.Rows; v++ {
			for h := 0; h < tiles.Columns; h++ {
				card := v*tiles.Columns + h
				if card >= maxcards {
					break card_loop
				}
				cards = append(cards, card)
				cards = append(cards, card)
			}
		}
	}

	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	cover = ebiten.NewImage(tiles.Width, tiles.Height)
	cover.Fill(borderColor)

	inset := ebiten.NewImage(tiles.Width-8, tiles.Height-8)
	inset.Fill(coverColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(4, 4)
	cover.DrawImage(inset, op)

	gw, gh = factors(len(cards))
	states = make([]bool, len(cards))

	canvas = ebiten.NewImage(tiles.Width*gw, tiles.Height*gh)
	canvas.Fill(coverColor)

	for ti, ci := range cards {
		x := ti % gw
		y := (ti / gw) % gh

		im := tiles.List[ci]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*tiles.Width), float64(y*tiles.Height))
		canvas.DrawImage(im, op)
	}

	moves = 0
	matches = 0
}

func factors(n int) (int, int) {
	for i := n - 1; i > 1; i-- {
		if n%i == 0 {
			m1 := i
			m2 := n / i

			if m2 >= m1 {
				return m2, m1
			}
		}
	}

	return n, 1
}

func main() {
	flag.IntVar(&maxcards, "cards", maxcards, "maximum number of cards")
	flag.DurationVar(&waitTurn, "turn", waitTurn, "wait before hiding cards (between turns)")
	flag.DurationVar(&waitGame, "game", waitGame, "wait before hiding cards (at game start)")
	audio := flag.Bool("audio", true, "play audio")
	flag.Parse()

	if maxcards > hcount*vcount {
		maxcards = hcount * vcount
	}

	rand.Seed(time.Now().Unix())

	initGame()

	if *audio {
		audioInit()
	}

	ebiten.SetWindowTitle("Memory")
	ebiten.SetWindowSize(gw*tiles.Width/2, gh*tiles.Height/2)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)

	if err := ebiten.RunGame(&Game{deal: make([]int, 0, 2)}); err != nil {
		log.Fatal(err)
	}

	fmt.Println(moves, "moves", matches, "matches")
}

func gameCoords(x, y int) (int, int) {
	return x / tiles.Width, y / tiles.Height
}

func gameIndex(x, y int) int {
	return y*gw + x
}

type gameState int

const (
	newGame gameState = iota
	waiting
	inGame
)

type Game struct {
	gameState gameState
	deal      []int
}

func (g *Game) frame() *ebiten.Image {
	if g.gameState != inGame {
		return canvas
	}

	bounds := canvas.Bounds()
	im := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	im.DrawImage(canvas, &ebiten.DrawImageOptions{})

	for ti, found := range states {
		if !found {
			x := ti % gw
			y := (ti / gw) % gh

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tiles.Width), float64(y*tiles.Height))
			im.DrawImage(cover, op)
		}
	}

	return im
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return gw * tiles.Width, gh * tiles.Height
}

func (g *Game) Draw(screen *ebiten.Image) {
	//op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.frame(), nil) //op)
}

func (g *Game) Update() error {
	switch {
	case g.gameState == newGame:
		g.gameState = waiting

		time.AfterFunc(waitGame, func() {
			g.gameState = inGame
		})

	case g.gameState == waiting:

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX):
		return fmt.Errorf("quit")

	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		initGame()
		g.deal = g.deal[:0]
		g.gameState = newGame

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		x, y := gameCoords(ebiten.CursorPosition())
		si := gameIndex(x, y)
		update := false

		if len(g.deal) < 2 && states[si] == false {
			states[si] = true
			g.deal = append(g.deal, si)

			if len(g.deal) == 2 {
				moves++
			}

			audioPlay(AudioFlip)
			update = true
		}

		if update && len(g.deal) == 2 {
			d1, d2 := g.deal[0], g.deal[1]

			if cards[d1] == cards[d2] {
				g.deal = g.deal[:0]
				matches++
			} else {
				time.AfterFunc(waitTurn, func() {
					states[d1] = false
					states[d2] = false
					g.deal = g.deal[:0]
					audioPlay(AudioReset)
				})
			}
		}

		if matches == maxcards {
			return fmt.Errorf("game over")
		}
	}

	return nil
}
