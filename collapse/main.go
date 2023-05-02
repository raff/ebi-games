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
)

const (
	hcount = 20
	vcount = 20
	border = 4
)

var (
	background = color.NRGBA{80, 80, 80, 255}
	//borderColor = color.NRGBA{160, 160, 160, 255}

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
	bw, bh int // window border
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

		//tw = (ww - border) / hcount
		//th = (wh - border) / vcount
		tw = ww / hcount
		th = wh / vcount

                ww += border
		wh += border

		//bw = (ww - (tw * hcount)) / 2
		//bh = (wh - (th * vcount)) / 2

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

	for y := 0; y < vcount+1; y++ {
		for x := 0; x < hcount; x++ {
			tile := ebiten.NewImage(tw-border, th-border)
			color := colors[rand.Intn(len(colors))]

			//fmt.Println(x,y, color)
			tile.Fill(color)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tw+border+bw), float64(y*th+border+bh))
			canvas.DrawImage(tile, op)
		}
	}

	return ww, wh
}

func main() {
	rand.Seed(time.Now().Unix())

	ww, wh = initCanvas(ebiten.ScreenSizeInFullscreen())

	ebiten.SetWindowTitle("Block Collapse")
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.RunGame(&Game{})
}

type Game struct{}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
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
