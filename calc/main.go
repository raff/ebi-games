package main

import (
	"bytes"
	"flag"
	"image"
	"image/png"
	"io"
	"log"

	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

//go:embed assets/calc.png
var calc_png []byte

func readImage(r io.Reader) (*ebiten.Image, int, int, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, 0, 0, err
	}

	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	ebimg := ebiten.NewImageFromImage(img)

	return ebimg, iw, ih, nil
}

func coords(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

const None = ebiten.Key(-2)

var buttons = map[ebiten.Key]image.Rectangle{
	ebiten.KeyC:     coords(9, 52, 18, 16),
	ebiten.KeyE:     coords(31, 52, 18, 16),
	ebiten.KeyEnter: coords(53, 52, 18, 16), // =
	ebiten.KeyX:     coords(75, 52, 18, 16), // *

	ebiten.Key7:     coords(9, 74, 18, 16),
	ebiten.Key8:     coords(31, 74, 18, 16),
	ebiten.Key9:     coords(53, 74, 18, 16),
	ebiten.KeySlash: coords(75, 74, 18, 16),

	ebiten.Key4:     coords(9, 96, 18, 16),
	ebiten.Key5:     coords(31, 96, 18, 16),
	ebiten.Key6:     coords(53, 96, 18, 16),
	ebiten.KeyMinus: coords(75, 96, 18, 16),

	ebiten.Key1:     coords(9, 118, 18, 16),
	ebiten.Key2:     coords(31, 118, 18, 16),
	ebiten.Key3:     coords(53, 118, 18, 16),
	ebiten.KeyEqual: coords(75, 118, 18, 38), // +

	ebiten.Key0:      coords(9, 140, 40, 16),
	ebiten.KeyPeriod: coords(53, 140, 18, 16),
}

var displayCoords = coords(13, 33, 77, 9) // display

type Game struct {
	redraw bool
	sel    ebiten.Key

	ima *ebiten.Image
	inv *ebiten.Image
	iw  int
	ih  int
	ww  int
	wh  int

	scale  int
	drawOp ebiten.DrawImageOptions
}

func newGame(scale int) *Game {
	ima, iw, ih, err := readImage(bytes.NewReader(calc_png))
	if err != nil {
		log.Fatal(err)
	}

	inv := ebiten.NewImage(iw, ih)

	var c colorm.ColorM
	c.Scale(-1, -1, -1, 1)
	c.Translate(1, 1, 1, 0)

	colorm.DrawImage(inv, ima, c, &colorm.DrawImageOptions{})

	sw, sh := ebiten.ScreenSizeInFullscreen()

	g := &Game{
		redraw: true,
		ima:    ima,
		inv:    inv,
		iw:     iw,
		ih:     ih,
		scale:  scale,
		ww:     iw,
		wh:     ih,
	}

	if scale != 1 {
		g.drawOp.GeoM.Scale(float64(scale), float64(scale))
		g.ww = g.ww * g.scale
		g.wh = g.wh * g.scale
	}

	if g.ww > sw || g.wh > sh {
		log.Fatal("scale too big")
	}

	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.ww, g.wh
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.redraw {
		return
	}

	screen.DrawImage(g.ima, &g.drawOp)

	if v, ok := buttons[g.sel]; ok {
		var drawOp ebiten.DrawImageOptions

		drawOp.GeoM.Scale(float64(g.scale), float64(g.scale))
		drawOp.GeoM.Translate(float64(v.Min.X*g.scale), float64(v.Min.Y*g.scale))

		inv := g.inv.SubImage(v).(*ebiten.Image)
		screen.DrawImage(inv, &drawOp)
	}

	g.redraw = false
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) { // (Q)uit or ESC
		return ebiten.Termination
	}

	p := image.Pt(ebiten.CursorPosition()).Div(g.scale)

	for k, v := range buttons {
		if ebiten.IsKeyPressed(k) || (ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && p.In(v)) {
			g.sel = k
			g.redraw = true
			return nil
		}
	}

	if g.sel != None {
		g.sel = None
		g.redraw = true
	}

	return nil
}

func main() {
	scale := flag.Int("scale", 4, "Window scale")
	flag.Parse()

	g := newGame(*scale)

	ebiten.SetWindowDecorated(false)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(g.ww, g.wh)
	ebiten.RunGame(g)
}
