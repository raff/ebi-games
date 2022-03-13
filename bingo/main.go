package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

const (
	hcount = 5
	vcount = 5
	border = 8
)

var (
	background  = color.NRGBA{80, 80, 80, 255}
	borderColor = color.NRGBA{160, 160, 160, 255}
	textColor   = color.NRGBA{0, 0, 0, 255}

	colors = []color.NRGBA{
		{255, 127, 0, 255},
		{255, 255, 0, 255},
		{0, 255, 0, 255},
		{64, 64, 255, 255},
		{255, 0, 255, 255},
	}

	canvas *ebiten.Image
	noop   = &ebiten.DrawImageOptions{}

	ffont font.Face

	tw, th int // game tile width, height
	ww, wh int // window width and height
)

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func initCanvas(w, h int) (int, int) {
	if w > 0 && h > 0 {
		ww, wh = w/2, h/2

		ww = min(ww, wh)
		wh = ww

		tw = (ww - border) / hcount
		th = (wh - border) / vcount

		wh += th

		canvas = ebiten.NewImage(ww, wh)
		canvas.Fill(background)

		tt, err := opentype.Parse(gobold.TTF)
		if err != nil {
			log.Fatal(err)
		}

		const dpi = 72
		ffont, err = opentype.NewFace(tt, &opentype.FaceOptions{
			Size:    48,
			DPI:     dpi,
			Hinting: font.HintingFull,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	tile := ebiten.NewImage(tw-border, th-border)
	tile.Fill(borderColor)

	inset := tile.SubImage(tile.Bounds().Inset(border)).(*ebiten.Image)

	var numbers []int

	for i := 0; i < 5; i++ {
		nn := rand.Perm(15)[:5]
		for _, n := range nn {
			numbers = append(numbers, n+1+(i*15))
		}
	}

	for y := 0; y < vcount+1; y++ {
		for x := 0; x < hcount; x++ {
			if true {
				inset.Fill(colors[x])

				var str string

				if y == 0 { // letters
					str = string("BINGO"[x])
				} else {
					i := x*hcount + y - 1
					str = fmt.Sprintf("%d", numbers[i]+1)
					if i == (vcount * hcount / 2) {
						str = "*"
					}
				}

				bound, _ := font.BoundString(ffont, str)
				w := (bound.Max.X - bound.Min.X).Ceil()
				h := (bound.Max.Y - bound.Min.Y).Ceil()
				x := (tile.Bounds().Dx() - w) / 2
				y := (tile.Bounds().Dy()-h)/2 + h
				text.Draw(inset, str, ffont, x, y, textColor)
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tw+border), float64(y*th+border))
			canvas.DrawImage(tile, op)
		}
	}

	return ww, wh
}

func main() {
	rand.Seed(time.Now().Unix())

	ww, wh = initCanvas(ebiten.ScreenSizeInFullscreen())

	ebiten.SetWindowTitle("Bingo")
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.RunGame(&Game{})
}

type Game struct{}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	//fmt.Println("layout", outsideWidth, outsideHeight, "-", ww, wh)
	return ww, wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(canvas, noop)
}

func (g *Game) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		initCanvas(0, 0)

	case inpututil.IsKeyJustPressed(ebiten.KeyQ), inpututil.IsKeyJustPressed(ebiten.KeyX):
		return fmt.Errorf("quit")
	}

	return nil
}
